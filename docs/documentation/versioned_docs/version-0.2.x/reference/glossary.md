---
sidebar_position: 1
---

# Glossary

This section explains the basic concepts that commonly mentioned in Permify, as well as in this document. You can find the whole context on right menu.

## Google Zanzibar (or just Zanzibar)

[Google Zanzibar] is the global authorization system used at Google for handling authorization for hundreds of its services and products including; YouTube, Drive, Calendar, Cloud and Maps.

Google published Zanzibar back in 2019, and in a short time it gained attention quickly. In fact some big tech companies started to shift their legacy authorization structure to Zanzibar style systems. Additionaly, Zanzibar based solutions increased over the time. All disclosure; [Permify] is an authorization system based on Zanzibar. 

For more about Zanzibar check our blog post, [Google Zanzibar In A Nutshell]

[Google Zanzibar In A Nutshell]: https://www.permify.co/post/google-zanzibar-in-a-nutshell
[Google Zanzibar]: https://research.google/pubs/pub48190/
[Permify]: https://www.permify.co/

## Permify Schema

Permify has its own language that you can model your authorization logic with it, we called it Permify Schema. The language allows to define arbitrary relations between users and objects, such as owner, editor, commenter or roles like admin, manager etc. You can define your entities, relations between them and access control decisions with using Permify Schema. 

It includes set-algebraic operators such as inter- section and union for specifying potentially complex access control policies in terms of those user-object relations.

## Relational Tuples

In Permify, relationship between your entities, objects, and users builds up a collection of access control lists (ACLs). 

These ACLs called relational tuples: the underlying data form that represents object-to-object and object-to-subject relations. The simplest form of relational tuple structured as `entity # relation @ user` and each relational tuple represents an action that a specific user or user set can do on a resource and takes form of `user U has relation R to object O`, where user U could be a simple user or a user set such as team X members.

## Write Database - WriteDB

Permify stores your relational tuples (authorization data) in **WriteDB**. You can configure it **WriteDB** when running Permify Service with using both [configuration flags](/docs/installation/brew#configuration-flags)  or [configuration YAML file](https://github.com/Permify/permify/blob/master/example.config.yaml).

## Relationship Based Access Control (ReBAC)

ReBAC is an access control model that defines permissions based on the relationships between entities and subjects of your system. Although ReBAC is best known for social networks because its core concept is about the network of relations, it can be applied beyond that. 

Check out [Relational Based Access Control Models](https://www.permify.co/post/relational-based-access-control-models) post learn more about ReBAC and its common usage.

## Domain Specific Language (DSL)

Domain Specific Language is a language that specialized to a particular application domain. Permify has its DSL basically an authorization language which you can model and structure your authorization with it. We called it Permify Schema.