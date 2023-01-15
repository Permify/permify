import React from "react";
import {Card} from "./Card"
import styles from "./Card.module.css";

export const CardList = () => {
  
  const list = [
    {
      id:1,
      title:"Try Permify in Local",
      description: "Set up Permify a with single docker command in your local",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212459030-7bd3ff7f-1538-4870-87cd-fbd0f4a21624.png",
      link: "./installation/overview"
    },
    {
      id:2,
      title:"Docker",
      description: "Deploy Permify on a server using a configuration yaml file",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458191-50464c53-3228-40bf-8e8c-66a021eac13a.svg",
      link: "./installation/container"
    },
    {
      id:3,
      title:"AWS",
      description: "Deploying Docker Container & Permify to AWS EC2 using ECS",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458359-e5472772-ce68-4c5a-a595-8ab123976202.svg",
      link:  "./installation/aws"
    },
    {
      id:4,
      title:"Kubernetes (EKS)",
      description: "Deploy Permify on a EKS Kubernetes cluster",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458403-4ad18c86-6618-4df4-86f6-167553fcee87.png",
      link:  "./installation/kubernetes"
    },
    {
      id:5,
      title:"Brew",
      description: "Install and run Permify with Brew",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458420-95a75de6-ac32-4958-87de-5e0848e6753c.png",
      link:  "./installation/brew"
    },
    {
      id:6,
      title:"Azure CR",
      description: "Deploy Permify with using Azure Container Registry",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458602-994075af-dbfd-408c-a8fb-d583653dce64.png",
      link: "./installation/azure"
    },
    {
      id:7,
      title:"Google Compute Engine",
      description: "Deploy Permify with using Google Compute Engine",
      imgSrc: "https://user-images.githubusercontent.com/34595361/212458849-354849d8-cbdf-48de-9272-6e6d9ad5856e.svg",
      link: "./installation/google"
    },
  ]

  return (
    <div className={styles["card-container-setup"]}>
      {list.map((item) => (
        <Card
          title={item.title}
          key={item.id}
          description={item.description}
          imgSrc={item.imgSrc}
          link={item.link}
        />
      ))}
    </div>
  );
};
