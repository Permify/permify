# Google Docs Style System

This example models a simplified version of Google Docs style permission system where users can be granted direct access to a resource, or access via organizations and nested groups.

### Model

```perm
entity user {}

entity resource {
    relation viewer  @user  @group#member @group#manager
    relation manager @user @group#member @group#manager
    
    action edit = manager
    action view = viewer or manager
}

entity group {
    relation manager @user @group#member @group#manager
    relation member @user @group#member @group#manager
}

entity organization {
    relation group @group
    relation resource @resource

    relation administrator @user @group#member @group#manager
    relation direct_member @user
   

    action admin = administrator
    action member = direct_member or administrator or group.member
}
```

## Breakdown of the Model 

### User 

```perm
entity user {}
```
Represents a user who can be granted permission to access a resource directly, or through their membership in a group or organization. 

### Resource 

```perm
entity resource {
    relation viewer  @user  @group#member @group#manager
    relation manager @user @group#member @group#manager

    action edit = manager
    action view = viewer or manager
}
```

Represents a resource that users can be granted permission to access. The resource entity has two relationships:

#### Relations 

**manager:** A relationship between users who are authorized to manage the resource. This relationship is defined by the `@user` annotation on both ends, and by the `@group#member` and `@group#manager` annotations on the ends corresponding to the group member and manager relations.

**viewer:** A relationship between users who are authorized to view the resource. This relationship is defined by the `@user` annotation on one end and the `@group#member` and `@group#manager` annotations on the other end corresponding to the group entity member and manager relations.

The resource entity has two actions defined:

#### Actions 

**manage:**: An action that can be performed by users who are authorized to manage the resource, as determined by the manager relationship.

**view:** An action that can be performed by users who are authorized to view the resource, as determined by the viewer and manager relationships.

### group

```perm
entity group {
    relation manager @user @group#member @group#manager
    relation member @user @group#member @group#manager
}
```

Represents a group of users who can be granted permission to access a resource. The group entity has two relationships:

#### Relations 

**manager:** A relationship between users who are authorized to manage the group. This relationship is defined by the `@user` annotation on both ends, and by the `@group#member` and `@group#manager` annotations on the ends corresponding to the group entity member and manager.

**member:** A relationship between users who are members of the group. This relationship is defined by the `@user` annotation on one end and the `@group#member` and `@group#manager` annotations on the other end corresponding to the group entity member and manager.

The group entity has one action defined:

### Organization 

```perm
entity organization {
    relation group @group
    relation resource @resource

    relation administrator @user @group#member @group#manager
    relation direct_member @user
   
    action admin = administrator
    action member = direct_member or administrator or group.member
}
```

Represents an organization that can contain groups, users, and resources. The organization entity has several relationships:

#### Relations 

**group:** A relationship between the organization and its groups. This relationship is defined by the `@group` annotation on the end corresponding to the group entity.

**administrator:** A relationship between users who are authorized to manage the organization. This relationship is defined by the `@user` annotation on both ends, and by the `@group#member` and `@group#manager` annotations on the ends corresponding to the group entity member and manager.

**direct_member:** A relationship between users who are directly members of the organization. This relationship is defined by the `@user` annotation on the end corresponding to the user entity.

**resource:** A relationship between the organization and its resources. This relationship is defined by the `@resource` annotation on the end corresponding to the resource entity.

The organization entity has two actions defined:

#### Actions 

**admin:** An action that can be performed by users who are authorized to manage the organization, as determined by the administrator relationship.

**member:** An action that can be performed by users who are directly members of the organization, or who have administrator relationship, or who are members of groups that are part of the organization, 

## Relationships

```perm
// Assign users to different groups
group:tech#manager@user:ashley
group:tech#member@user:david
group:marketing#manager@user:john
group:marketing#member@user:jenny
group:hr#manager@user:josh
group:hr#member@user:joe

// Assign groups to other groups
group:tech#member@group:marketing#member
group:tech#member@group:hr#member

// Connect groups to organization.
organization:acme#group@group:tech
organization:acme#group@group:marketing
organization:acme#group@group:hr

// Add some resources under the organization
organization:acme#resource@resource:product_database
organization:acme#resource@resource:marketing_materials
organization:acme#resource@resource:hr_documents

// Assign a user and members of a group as administrators for the organization
organization:acme#administrator@group:tech#manager
organization:acme#administrator@user:it_admin

// Set the permissions on some resources
resource:product_database#manager@group:tech#manager
resource:product_database#viewer@group:tech#member
resource:marketing_materials#viewer@group:marketing#member
resource:hr_documents#manager@group:hr#manager
resource:hr_documents#viewer@group:hr#member
```

## See on the Playground

Here is the visualization of the relationships of the schema:

![visualization](https://user-images.githubusercontent.com/34595361/231216456-1430d952-856a-4dad-996b-968a1a59fc04.png)

You can see and play around with this example in our playground using this [link](https://play.permify.co/?s=iuRic3nR1HeZJcFyRNKPo).

