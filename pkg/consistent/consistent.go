package consistent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
)

const (
	// DefaultPartitionCount defines the default number of virtual partitions in the hash ring.
	// This helps balance the load distribution among members, even with a small number of members.
	DefaultPartitionCount int = 271

	// DefaultReplicationFactor specifies the default number of replicas for each partition.
	// This ensures redundancy and fault tolerance by assigning partitions to multiple members.
	DefaultReplicationFactor int = 20

	// DefaultLoad defines the default maximum load factor for each member.
	// A higher value allows members to handle more load before being considered full.
	DefaultLoad float64 = 1.25

	// DefaultPickerWidth determines the default range of candidates considered when picking members.
	// This can influence the selection logic in advanced configurations.
	DefaultPickerWidth int = 1
)

type Hasher func([]byte) uint64

type Member interface {
	String() string
}

// Config represents the configuration settings for a specific system or application.
// It includes settings for hashing, partitioning, replication, load balancing, and picker width.
type Config struct {
	// Hasher is an interface or implementation used for generating hash values.
	// It is typically used to distribute data evenly across partitions.
	Hasher Hasher

	// PartitionCount defines the number of partitions in the system.
	// This value affects how data is distributed and processed.
	PartitionCount int

	// ReplicationFactor specifies the number of replicas for each partition.
	// It ensures data redundancy and fault tolerance in the system.
	ReplicationFactor int

	// Load represents the load balancing factor for the system.
	// It could be a threshold or weight used for distributing work.
	Load float64

	// PickerWidth determines the width or range of the picker mechanism.
	// It is typically used to influence how selections are made in certain operations.
	PickerWidth int
}

// Consistent implements a consistent hashing mechanism with partitioning and load balancing.
// It is used for distributing data across a dynamic set of members efficiently.
type Consistent struct {
	// mu is a read-write mutex used to protect shared resources from concurrent access.
	mu sync.RWMutex

	// config holds the configuration settings for the consistent hashing instance.
	config Config

	// hasher is an implementation of the Hasher interface used for generating hash values.
	hasher Hasher

	// sortedSet maintains a sorted slice of hash values to represent the hash ring.
	sortedSet []uint64

	// partitionCount specifies the number of partitions in the hash ring.
	partitionCount uint64

	// loads tracks the load distribution for each member in the hash ring.
	// The key is the member's identifier, and the value is the load.
	loads map[string]float64

	// members is a map of member identifiers to their corresponding Member struct.
	members map[string]*Member

	// partitions maps each partition index to the corresponding member.
	partitions map[int]*Member

	// ring is a map that associates each hash value in the ring with a specific member.
	ring map[uint64]*Member
}

// New initializes and returns a new instance of the Consistent struct.
// It takes a Config parameter and applies default values for any unset fields.
func New(config Config) *Consistent {
	// Ensure the Hasher is not nil; a nil Hasher would make consistent hashing unusable.
	if config.Hasher == nil {
		panic("Hasher cannot be nil")
	}

	// Set default values for partition count, replication factor, load, and picker width if not provided.
	if config.PartitionCount == 0 {
		config.PartitionCount = DefaultPartitionCount
	}
	if config.ReplicationFactor == 0 {
		config.ReplicationFactor = DefaultReplicationFactor
	}
	if config.Load == 0 {
		config.Load = DefaultLoad
	}
	if config.PickerWidth == 0 {
		config.PickerWidth = DefaultPickerWidth
	}

	// Initialize a new Consistent instance with the provided configuration.
	c := &Consistent{
		config:         config,
		members:        make(map[string]*Member),
		partitionCount: uint64(config.PartitionCount),
		ring:           make(map[uint64]*Member),
	}

	// Assign the provided Hasher implementation to the instance.
	c.hasher = config.Hasher
	return c
}

// Members returns a slice of all the members currently in the consistent hash ring.
// It safely retrieves the members using a read lock to prevent data races while
// accessing the shared `members` map.
func (c *Consistent) Members() []Member {
	// Acquire a read lock to ensure thread-safe access to the members map.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a slice to hold the members, pre-allocating its capacity to avoid resizing.
	members := make([]Member, 0, len(c.members))

	// Iterate over the members map and append each member to the slice.
	for _, member := range c.members {
		members = append(members, *member)
	}

	// Return the slice of members.
	return members
}

// GetAverageLoad calculates and returns the current average load across all members.
// It is a public method that provides thread-safe access to the load calculation.
func (c *Consistent) GetAverageLoad() float64 {
	// Acquire a read lock to ensure thread-safe access to shared resources.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Delegate the actual load calculation to the internal helper method.
	return c.calculateAverageLoad()
}

// calculateAverageLoad is a private helper method that performs the actual calculation
// of the average load across all members. It is not thread-safe and should be called
// only from within methods that already manage locking.
func (c *Consistent) calculateAverageLoad() float64 {
	// If there are no members, return an average load of 0 to prevent division by zero.
	if len(c.members) == 0 {
		return 0
	}

	// Calculate the average load by dividing the total partition count by the number of members
	// and multiplying by the configured load factor.
	avgLoad := float64(c.partitionCount/uint64(len(c.members))) * c.config.Load

	// Use math.Ceil to round up the average load to the nearest whole number.
	return math.Ceil(avgLoad)
}

// assignPartitionWithLoad distributes a partition to a member based on the load factor.
// It ensures that no member exceeds the calculated average load while distributing partitions.
// If the distribution fails due to insufficient capacity, it panics with an error message.
func (c *Consistent) assignPartitionWithLoad(
	partitionID, startIndex int,
	partitionAssignments map[int]*Member,
	memberLoads map[string]float64,
) {
	// Calculate the average load to determine the maximum load a member can handle.
	averageLoad := c.calculateAverageLoad()
	var attempts int

	// Iterate to find a suitable member for the partition.
	for {
		attempts++

		// If the loop exceeds the number of members, it indicates that the partition
		// cannot be distributed with the current configuration.
		if attempts >= len(c.sortedSet) {
			panic("not enough capacity to distribute partitions: consider decreasing the partition count, increasing the member count, or increasing the load factor")
		}

		// Get the current hash value from the sorted set.
		currentHash := c.sortedSet[startIndex]

		// Retrieve the member associated with the hash value.
		currentMember := *c.ring[currentHash]

		// Check the current load of the member.
		currentLoad := memberLoads[currentMember.String()]

		// If the member's load is within the acceptable range, assign the partition.
		if currentLoad+1 <= averageLoad {
			partitionAssignments[partitionID] = &currentMember
			memberLoads[currentMember.String()]++
			return
		}

		// Move to the next member in the sorted set.
		startIndex++
		if startIndex >= len(c.sortedSet) {
			// Loop back to the beginning of the sorted set if we reach the end.
			startIndex = 0
		}
	}
}

// distributePartitions evenly distributes partitions among members while respecting the load factor.
// It ensures that partitions are assigned to members based on consistent hashing and load constraints.
func (c *Consistent) distributePartitions() {
	// Initialize maps to track the load for each member and partition assignments.
	memberLoads := make(map[string]float64)
	partitionAssignments := make(map[int]*Member)

	// Create a buffer for converting partition IDs into byte slices for hashing.
	partitionKeyBuffer := make([]byte, 8)

	// Iterate over all partition IDs to distribute them among members.
	for partitionID := uint64(0); partitionID < c.partitionCount; partitionID++ {
		// Convert the partition ID into a byte slice for hashing.
		binary.LittleEndian.PutUint64(partitionKeyBuffer, partitionID)

		// Generate a hash key for the partition using the configured hasher.
		hashKey := c.hasher(partitionKeyBuffer)

		// Find the index of the member in the sorted set where the hash key should be placed.
		index := sort.Search(len(c.sortedSet), func(i int) bool {
			return c.sortedSet[i] >= hashKey
		})

		// If the index is beyond the end of the sorted set, wrap around to the beginning.
		if index >= len(c.sortedSet) {
			index = 0
		}

		// Assign the partition to a member, ensuring the load factor is respected.
		c.assignPartitionWithLoad(int(partitionID), index, partitionAssignments, memberLoads)
	}

	// Update the Consistent instance with the new partition assignments and member loads.
	c.partitions = partitionAssignments
	c.loads = memberLoads
}

// addMemberToRing adds a member to the consistent hash ring and updates the sorted set of hashes.
func (c *Consistent) addMemberToRing(member Member) {
	// Add replication factor entries for the member in the hash ring.
	for replicaIndex := 0; replicaIndex < c.config.ReplicationFactor; replicaIndex++ {
		// Generate a unique key for each replica of the member.
		replicaKey := []byte(fmt.Sprintf("%s%d", member.String(), replicaIndex))
		hashValue := c.hasher(replicaKey)

		// Add the hash value to the ring and associate it with the member.
		c.ring[hashValue] = &member

		// Append the hash value to the sorted set of hashes.
		c.sortedSet = append(c.sortedSet, hashValue)
	}

	// Sort the hash values to maintain the ring's order.
	sort.Slice(c.sortedSet, func(i, j int) bool {
		return c.sortedSet[i] < c.sortedSet[j]
	})

	// Add the member to the members map.
	c.members[member.String()] = &member
}

// Add safely adds a new member to the consistent hash circle.
// It ensures thread safety and redistributes partitions after adding the member.
func (c *Consistent) Add(member Member) {
	// Acquire a write lock to ensure thread safety.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the member already exists in the ring. If it does, exit early.
	if _, exists := c.members[member.String()]; exists {
		return
	}

	// Add the member to the ring and redistribute partitions.
	c.addMemberToRing(member)
	c.distributePartitions()
}

// removeFromSortedSet removes a hash value from the sorted set of hashes.
func (c *Consistent) removeFromSortedSet(hashValue uint64) {
	for i := 0; i < len(c.sortedSet); i++ {
		if c.sortedSet[i] == hashValue {
			// Remove the hash value by slicing the sorted set.
			c.sortedSet = append(c.sortedSet[:i], c.sortedSet[i+1:]...)
			break
		}
	}
}

// Remove deletes a member from the consistent hash circle and redistributes partitions.
// If the member does not exist, the method exits early.
func (c *Consistent) Remove(memberName string) {
	// Acquire a write lock to ensure thread-safe access.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the member exists in the hash ring. If not, exit early.
	if _, exists := c.members[memberName]; !exists {
		return
	}

	// Remove all replicas of the member from the hash ring and sorted set.
	for replicaIndex := 0; replicaIndex < c.config.ReplicationFactor; replicaIndex++ {
		// Generate the unique key for each replica of the member.
		replicaKey := []byte(fmt.Sprintf("%s%d", memberName, replicaIndex))
		hashValue := c.hasher(replicaKey)

		// Remove the hash value from the hash ring.
		delete(c.ring, hashValue)

		// Remove the hash value from the sorted set.
		c.removeFromSortedSet(hashValue)
	}

	// Remove the member from the members map.
	delete(c.members, memberName)

	// If no members remain, reset the partition table and exit.
	if len(c.members) == 0 {
		c.partitions = make(map[int]*Member)
		return
	}

	// Redistribute partitions among the remaining members.
	c.distributePartitions()
}

// GetLoadDistribution provides a thread-safe snapshot of the current load distribution across members.
// It returns a map where the keys are member identifiers and the values are their respective loads.
func (c *Consistent) GetLoadDistribution() map[string]float64 {
	// Acquire a read lock to ensure thread-safe access to the loads map.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy of the loads map to avoid exposing internal state.
	loadDistribution := make(map[string]float64)
	for memberName, memberLoad := range c.loads {
		loadDistribution[memberName] = memberLoad
	}

	return loadDistribution
}

// GetPartitionID calculates and returns the partition ID for a given key.
// The partition ID is determined by hashing the key and applying modulo operation with the partition count.
func (c *Consistent) GetPartitionID(key []byte) int {
	// Generate a hash value for the given key using the configured hasher.
	hashValue := c.hasher(key)

	// Calculate the partition ID by taking the modulus of the hash value with the partition count.
	return int(hashValue % c.partitionCount)
}

// GetPartitionOwner retrieves the owner of the specified partition in a thread-safe manner.
// It ensures that the access to shared resources is synchronized.
func (c *Consistent) GetPartitionOwner(partitionID int) Member {
	// Acquire a read lock to ensure thread-safe access to the partitions map.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Delegate the actual lookup to the non-thread-safe internal helper function.
	return c.getPartitionOwnerInternal(partitionID)
}

// getPartitionOwnerInternal retrieves the owner of the specified partition without thread safety.
// This function assumes that synchronization has been handled by the caller.
func (c *Consistent) getPartitionOwnerInternal(partitionID int) Member {
	// Lookup the member associated with the given partition ID.
	member, exists := c.partitions[partitionID]
	if !exists {
		// If the partition ID does not exist, return a nil Member.
		return nil
	}

	// Return a copy of the member to ensure thread safety.
	return *member
}

// LocateKey determines the owner of the partition corresponding to the given key.
// It calculates the partition ID for the key and retrieves the associated member in a thread-safe manner.
func (c *Consistent) LocateKey(key []byte) Member {
	// Calculate the partition ID based on the hash of the key.
	partitionID := c.GetPartitionID(key)

	// Retrieve the owner of the partition using the thread-safe method.
	return c.GetPartitionOwner(partitionID)
}

// closestN retrieves the closest N members to the given partition ID in the consistent hash ring.
// It ensures thread-safe access and validates that the requested count of members can be satisfied.
func (c *Consistent) closestN(partitionID, count int) ([]Member, error) {
	// Acquire a read lock to ensure thread-safe access to the members map.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Validate that the requested number of members can be satisfied.
	if count > len(c.members) {
		return nil, errors.New("not enough members to satisfy the request")
	}

	// Prepare a result slice to store the closest members.
	var closestMembers []Member

	// Get the owner of the given partition.
	partitionOwner := c.getPartitionOwnerInternal(partitionID)
	var partitionOwnerHash uint64

	// Build a hash ring by hashing all member names.
	var memberHashes []uint64
	hashToMember := make(map[uint64]*Member)
	for memberName, member := range c.members {
		// Compute the hash value for each member name.
		hash := c.hasher([]byte(memberName))

		// Track the hash for the partition owner.
		if memberName == partitionOwner.String() {
			partitionOwnerHash = hash
		}

		// Append the hash value and map it to the corresponding member.
		memberHashes = append(memberHashes, hash)
		hashToMember[hash] = member
	}

	// Sort the hash values to create a consistent hash ring.
	sort.Slice(memberHashes, func(i, j int) bool {
		return memberHashes[i] < memberHashes[j]
	})

	// Find the index of the partition owner's hash in the sorted hash ring.
	ownerIndex := -1
	for i, hash := range memberHashes {
		if hash == partitionOwnerHash {
			ownerIndex = i
			closestMembers = append(closestMembers, *hashToMember[hash])
			break
		}
	}

	// If the partition owner's hash is not found (unexpected), return an error.
	if ownerIndex == -1 {
		return nil, errors.New("partition owner not found in hash ring")
	}

	// Find the additional closest members by iterating around the hash ring.
	currentIndex := ownerIndex
	for len(closestMembers) < count {
		// Move to the next hash in the ring, wrapping around if necessary.
		currentIndex++
		if currentIndex >= len(memberHashes) {
			currentIndex = 0
		}

		// Add the member corresponding to the current hash to the result.
		hash := memberHashes[currentIndex]
		closestMembers = append(closestMembers, *hashToMember[hash])
	}

	return closestMembers, nil
}

// ClosestN calculates the closest N members to a given key in the consistent hash ring.
// It uses the key to determine the partition ID and then retrieves the closest members.
// This is useful for identifying members for replication or redundancy.
func (c *Consistent) ClosestN(key []byte, count int) ([]Member, error) {
	// Calculate the partition ID based on the hash of the key.
	partitionID := c.GetPartitionID(key)

	// Retrieve the closest N members for the calculated partition ID.
	return c.closestN(partitionID, count)
}
