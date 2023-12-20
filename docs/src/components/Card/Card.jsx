import React from "react";
import styles from "./Card.module.css";

export const Card = ({
  title,
  description,
  imgSrc,
  link,
}) => {
  console.log('link:', link)

  return (
    <div>
      <a
        href={link}
        className={styles["card"]}
        style={{ textDecoration: "none", color: "inherit" }}
      >
      <div className={styles["card-body"]}>
        <div className={styles["card-icon"]}>
          <img className="img" src={imgSrc} width="100%" />
        </div>
        <div className={styles["card-info"]}>
          <h3 style={{ margin: "0", paddingBottom: "0.5rem" }}>{title}</h3>
          <p style={{ margin: "0"}}>{description}</p>
        </div>
      </div>
    </a>
  </div>
  );
};
