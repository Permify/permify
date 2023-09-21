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
      title:"Relationship Based Access Control (ReBAC)",
      description: "Define permissions based on the relationships between resources and subjects in your system",
      link: "./use-cases/rebac"
    },
    {
      id:4,
      title:"Custom Roles",
      description: "Assign specific permissions to users based on the custom roles that they are assigned within the system.",
      link: "./use-cases/custom-roles"
    },
    {
      id:5,
      title:"Multi Tenancy",
      description: "Create custom authorization schema and relation tuples for the different tenants and manage them in a single place.",
      link:  "./use-cases/multi-tenancy"
    },
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
