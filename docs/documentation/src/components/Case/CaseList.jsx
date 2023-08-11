import React from "react";
import {Case} from "./Case"
import styles from "./Case.module.css";

export const CaseList = () => {
  
  const list = [
    {
      id:1,
      title:"Role Based Access Control (RBAC)",
      description: "Want to implement role to your application ? Define an entity and manage your roles throught your applications.",
      link: "./use-cases/simple-rbac"
    },
    {
      id:2,
      title:"Attribute Based Access Control (ABAC)",
      description: "Grant access what based on specific characteristics or attributes.",
      link: "./use-cases/abac"
    },
    {
      id:3,
      title:"Nested Hierarchies (ReBAC)",
      description: "Define nested parent child relationships to control access of your resources and inherit/share permissions between your entites.",
      link: "./use-cases/nested-hierarchies"
    },
    {
      id:4,
      title:"Organization Specific Resources",
      description: "Grant and manage user access to the organizational-wide resources; files, folders, repositories, etc.",
      link: "./use-cases/organizational"
    },
    {
      id:5,
      title:"Custom Roles",
      description: "Assign specific permissions to users based on the custom roles that they are assigned within the system.",
      link: "./use-cases/custom-roles"
    },
    {
      id:6,
      title:"Multi Tenancy",
      description: "Create custom authorization schema and relation tuples for the different tenants and manage them in a single place.",
      link:  "./multi-tenancy"
    },
    {
      id:7,
      title:"User Groups & Teams",
      description: "Grant permissions and privileges to the users according to the group or team that they belongs to.",
      link: "./use-cases/user-groups"
    },
    {
      id:8,
      title:"Sharing and Collaboration",
      description: "Invite a user or colleague to a resource and manage permissions accordingly.",
      link: "./use-cases/sharing"
    }
  ]

  return (
    <div className={styles["card-container-setup"]}>
      {list.map((item) => (
        <Case
          title={item.title}
          key={item.id}
          description={item.description}
          link={item.link}
        />
      ))}
    </div>
  );
};
