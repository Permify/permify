# Mercury

Explore **Mercury's Authorization Schema** in this example, delving into the intricate interplay among **users**, **organizations**, and **accounts**. Uncover the defined user roles, approval workflows, and limits, providing a snapshot of the dynamic relationships within the Mercury ecosystem.

## Schema | [Open in playground](https://play.permify.co/?s=mercury&tab=schema)

```perm
entity user {}

entity organization {
    relation admin @user
    relation member @user

    attribute admin_approval_limit integer
    attribute member_approval_limit integer
    attribute approval_num integer

    action approve = admin
    action create_account = admin

    permission approval = (member and check_member_approval(approval_num, member_approval_limit)) or (admin and check_admin_approval(approval_num, admin_approval_limit))
}

entity account {
    relation checkings @account
    relation savings @account

    relation owner @organization

    attribute withdraw_limit double
    attribute balance double

    action withdraw = check_balance(balance, request.amount) and (check_limit(withdraw_limit, request.amount) or owner.approval)
}

rule check_balance(balance double, amount double) {
    balance >= amount
}

rule check_limit(withdraw_limit double, amount double) {
    withdraw_limit >= amount
}

rule check_admin_approval(approval_num integer, admin_approval_limit integer) {
    approval_num >= admin_approval_limit
}

rule check_member_approval(approval_num integer, member_approval_limit integer) {
    approval_num >= member_approval_limit
}
```

## Brief Examination of the Model

Mercury's authorization model consists of three primary entities: **user**, **organization**, and **account**.
These entities are interconnected through defined relations and governed by specific rules and actions.

### Entities & Relations

**user**: Represents individual users within the system.

**organization**: Represents organizational entities and establishes relations with users (`admin` and `member`). Additionally, this entity holds attributes like `admin_approval_limit`, `member_approval_limit`, and `approval_num`.

**account**: Represents user accounts with relations to different account types (`checkings` and `savings`). It also has a relation to the owning `organization` and attributes such as `withdraw_limit` and `balance`.

### Permissions

The authorization schema defines two crucial permissions:

**approval**: Determines the conditions under which a user (either `member` or `admin`) can approve actions based on approval limits.

## Need any help ?

This is the end of demonstration of the authorization structure for Facebook groups. To install and implement this see the [Set Up Permify](../../installation.md) section.
