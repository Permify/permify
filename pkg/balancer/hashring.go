package balancer

import (
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

const (
	// Policy defines the name or identifier for the consistent hashing load balancing policy.
	Policy = "consistenthashpolicy"

	// Key is the context key used to retrieve the hash key for consistent hashing.
	Key = "consistenthashkey"

	// ConnectionLifetime specifies the duration for which a connection is maintained
	// before being considered for termination or renewal.
	ConnectionLifetime = time.Second * 5
)

// subConnInfo records the state and addr corresponding to the SubConn.
type subConnInfo struct {
	state connectivity.State
	addr  string
}

type consistentHashBalancer struct {
	// clientConn represents the client connection that created this balancer.
	clientConn balancer.ClientConn

	// connectionState indicates the current state of the connection.
	connectionState connectivity.State

	// addressInfoMap maps address strings to their corresponding resolver addresses.
	addressInfoMap map[string]resolver.Address

	// subConnectionMap maps address strings to their associated sub-connections.
	subConnectionMap map[string]balancer.SubConn

	// subConnInfoSyncMap is a concurrent map that stores information about each sub-connection.
	subConnInfoSyncMap sync.Map

	// currentPicker represents the picker currently used by the client connection for load balancing.
	currentPicker balancer.Picker

	// lastResolverError stores the most recent error reported by the resolver.
	lastResolverError error

	// lastConnectionError stores the most recent connection-related error.
	lastConnectionError error

	// pickerResultChannel is a channel through which pickers report their results.
	pickerResultChannel chan PickResult

	// activePickResults is a queue for storing pick results that have active contexts.
	activePickResults *Queue

	// subConnPickCounts tracks the number of pick results associated with each sub-connection.
	subConnPickCounts map[balancer.SubConn]*int32

	// subConnStatusMap indicates the status (active/inactive) of each sub-connection.
	subConnStatusMap map[balancer.SubConn]bool

	// balancerLock is a mutex used to ensure thread safety, especially when accessing the subConnPickCounts map.
	balancerLock sync.Mutex
}

// UpdateClientConnState processes the provided ClientConnState and updates
// the internal state of the balancer accordingly.
func (b *consistentHashBalancer) UpdateClientConnState(s balancer.ClientConnState) error {
	b.balancerLock.Lock() // Ensure exclusive access to balancers data.
	defer b.balancerLock.Unlock()

	// Update address information and get a set of active addresses.
	addrsSet := b.updateAddressInfo(s)

	// Remove any sub-connections that are no longer active.
	b.removeStaleSubConns(addrsSet)

	// If there are no addresses from the resolver, log an error.
	if len(s.ResolverState.Addresses) == 0 {
		b.ResolverError(fmt.Errorf("produced zero addresses"))
		return balancer.ErrBadResolverState
	}

	// Re-generate the picker based on the updated state and inform the client connection.
	b.regeneratePicker()
	b.clientConn.UpdateState(balancer.State{ConnectivityState: b.connectionState, Picker: b.currentPicker})

	return nil
}

// updateAddressInfo processes the provided ClientConnState and updates
// the balancers address information. It returns a set of active addresses.
func (b *consistentHashBalancer) updateAddressInfo(s balancer.ClientConnState) map[string]struct{} {
	addrsSet := make(map[string]struct{})

	// Iterate over the addresses from the resolver.
	for _, a := range s.ResolverState.Addresses {
		addr := a.Addr
		b.addressInfoMap[addr] = a
		addrsSet[addr] = struct{}{}

		// If there isn't a sub-connection for this address, create one.
		if sc, ok := b.subConnectionMap[addr]; !ok {
			if err := b.createNewSubConn(a, addr); err != nil {
				log.Printf("Consistent Hash Balancer: failed to create new SubConn: %v", err)
			}
		} else {
			// If a sub-connection exists, update its addresses.
			b.clientConn.UpdateAddresses(sc, []resolver.Address{a})
		}
	}

	return addrsSet
}

// createNewSubConn creates a new sub-connection for the provided address.
func (b *consistentHashBalancer) createNewSubConn(a resolver.Address, addr string) error {
	newSC, err := b.clientConn.NewSubConn([]resolver.Address{a}, balancer.NewSubConnOptions{HealthCheckEnabled: false})
	if err != nil {
		return err
	}

	// Store the new sub-connection and its info.
	b.subConnectionMap[addr] = newSC
	b.subConnInfoSyncMap.Store(newSC, &subConnInfo{
		state: connectivity.Idle,
		addr:  addr,
	})
	newSC.Connect()

	return nil
}

// removeStaleSubConns removes sub-connections that are no longer in the active addresses set.
func (b *consistentHashBalancer) removeStaleSubConns(addrsSet map[string]struct{}) {
	for a, sc := range b.subConnectionMap {
		// If a sub-connection's address isn't in the active set, remove it.
		if _, ok := addrsSet[a]; !ok {
			b.clientConn.RemoveSubConn(sc)
			delete(b.subConnectionMap, a)
			b.subConnInfoSyncMap.Delete(sc) // Cleanup related data.
		}
	}
}

// ResolverError handles resolver errors and updates state/picker accordingly.
func (b *consistentHashBalancer) ResolverError(err error) {
	b.balancerLock.Lock() // Ensure exclusive access to balancers data.
	defer b.balancerLock.Unlock()

	// Store the error and re-generate the picker.
	b.lastResolverError = err
	b.regeneratePicker()

	// Update client connection state if in a TransientFailure state.
	if b.connectionState != connectivity.TransientFailure {
		return
	}
	b.clientConn.UpdateState(balancer.State{
		ConnectivityState: b.connectionState,
		Picker:            b.currentPicker,
	})
}

// regeneratePicker generates a new picker to replace the old one with new data, and update the state of the balancer.
func (b *consistentHashBalancer) regeneratePicker() {
	availableSCs := make(map[string]balancer.SubConn)

	for addr, sc := range b.subConnectionMap {
		if stIface, ok := b.subConnInfoSyncMap.Load(sc); ok {
			if st, ok := stIface.(*subConnInfo); ok {
				// Only include sub-connections that are in a Ready or Idle state
				if st.state == connectivity.Ready || st.state == connectivity.Idle {
					availableSCs[addr] = sc
				}
			} else {
				log.Printf("Unexpected type in scInfos for key %v: expected *subConnInfo, got %T", sc, stIface)
			}
		}
	}

	if len(availableSCs) == 0 {
		b.connectionState = connectivity.TransientFailure
		b.currentPicker = base.NewErrPicker(b.mergeErrors())
	} else {
		b.connectionState = connectivity.Ready
		b.currentPicker = NewConsistentHashPicker(availableSCs)
	}
}

// mergeErrors -
func (b *consistentHashBalancer) mergeErrors() error {
	// If both errors are nil, return a generic error.
	if b.lastConnectionError == nil && b.lastResolverError == nil {
		return fmt.Errorf("unknown error occurred")
	}

	// If only one of the errors is nil, return the other error.
	if b.lastConnectionError == nil {
		return fmt.Errorf("last resolver error: %v", b.lastResolverError)
	}
	if b.lastResolverError == nil {
		return fmt.Errorf("last connection error: %v", b.lastConnectionError)
	}

	// If both errors are present, concatenate them.
	return fmt.Errorf("last connection error: %v; last resolver error: %v", b.lastConnectionError, b.lastResolverError)
}

// UpdateSubConnState -
func (b *consistentHashBalancer) UpdateSubConnState(subConn balancer.SubConn, stateUpdate balancer.SubConnState) {
	currentState := stateUpdate.ConnectivityState

	storedInfo, infoExists := b.subConnInfoSyncMap.Load(subConn)
	if !infoExists {
		// If the subConn isn't in our info map, it's a no-op for us.
		return
	}

	subConnInfo := storedInfo.(*subConnInfo)
	previousState := subConnInfo.state

	if previousState == currentState {
		// If the state hasn't changed, no need for further processing.
		return
	}

	slog.Debug("State of one sub-connection changed", slog.String("previous", previousState.String()), slog.String("current", currentState.String()))

	// Handle transitions from TransientFailure to Connecting.
	if previousState == connectivity.TransientFailure && currentState == connectivity.Connecting {
		return
	}

	// Update the state in the stored sub-connection info.
	subConnInfo.state = currentState

	switch currentState {
	case connectivity.Idle:
		subConn.Connect()
	case connectivity.Shutdown:
		b.subConnInfoSyncMap.Delete(subConn)
	case connectivity.TransientFailure:
		b.lastConnectionError = stateUpdate.ConnectionError
	}

	// If there's a significant change in the connection state, regenerate the picker.
	if hasSignificantStateChange(previousState, currentState) || b.connectionState == connectivity.TransientFailure {
		b.regeneratePicker()
	}

	b.clientConn.UpdateState(balancer.State{ConnectivityState: b.connectionState, Picker: b.currentPicker})
}

// hasSignificantStateChange -
func hasSignificantStateChange(oldState, newState connectivity.State) bool {
	isOldStateSignificant := oldState == connectivity.TransientFailure || oldState == connectivity.Shutdown
	isNewStateSignificant := newState == connectivity.TransientFailure || newState == connectivity.Shutdown
	return isOldStateSignificant != isNewStateSignificant
}

// Close -
func (b *consistentHashBalancer) Close() {}

// resetSubConn resets the given SubConn.
// It first retrieves the address associated with the SubConn, then tries to reset it using that address.
func (b *consistentHashBalancer) resetSubConn(subConn balancer.SubConn) error {
	// Get the address associated with the SubConn.
	address, err := b.getSubConnAddr(subConn)
	if err != nil {
		return fmt.Errorf("failed to get address for sub connection: %v", err)
	}

	slog.Debug("Resetting connection with", slog.String("address", address))

	// Reset the SubConn using its address.
	if resetErr := b.resetSubConnWithAddr(address); resetErr != nil {
		return fmt.Errorf("failed to reset sub connection with address %s: %v", address, resetErr)
	}

	return nil
}

// getSubConnAddr retrieves the address associated with the given SubConn from the stored connection info.
func (b *consistentHashBalancer) getSubConnAddr(subConn balancer.SubConn) (string, error) {
	// Load the sub connection info from the sync map.
	connInfoValue, exists := b.subConnInfoSyncMap.Load(subConn)

	if !exists {
		return "", ErrSubConnMissing
	}

	// Type assert and return the address from the sub connection info.
	subConnInfo := connInfoValue.(*subConnInfo)
	return subConnInfo.addr, nil
}

// resetSubConnWithAddr replaces the current SubConn associated with the provided address with a new one.
func (b *consistentHashBalancer) resetSubConnWithAddr(address string) error {
	// Retrieve the current SubConn associated with the address.
	currentSubConn, exists := b.subConnectionMap[address]
	if !exists {
		return ErrSubConnMissing
	}

	// Delete the info and remove the SubConn.
	b.subConnInfoSyncMap.Delete(currentSubConn)
	b.clientConn.RemoveSubConn(currentSubConn)

	// Fetch the address information.
	addressInfo, infoExists := b.addressInfoMap[address]
	if !infoExists {
		slog.Error("Consistent Hash Balancer: Address information missing for", slog.String("address", address))
		return ErrSubConnResetFailure
	}

	// Create a new SubConn with the address information.
	newSubConn, err := b.clientConn.NewSubConn([]resolver.Address{addressInfo}, balancer.NewSubConnOptions{HealthCheckEnabled: false})
	if err != nil {
		return err
	}

	// Store the new SubConn and its information.
	b.subConnectionMap[address] = newSubConn
	b.subConnInfoSyncMap.Store(newSubConn, &subConnInfo{
		state: connectivity.Idle,
		addr:  address,
	})

	// Regenerate picker and update the balancer state.
	b.regeneratePicker()
	b.clientConn.UpdateState(balancer.State{ConnectivityState: b.connectionState, Picker: b.currentPicker})

	return nil
}
