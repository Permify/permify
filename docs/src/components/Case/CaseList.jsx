import React from "react";
import { Case } from "./Case";
import styles from "./Case.module.css";

export const CaseList = ({ list }) => {
  return (
    <div className={styles["card-container-setup"]}>
      {list && list.map((item) => (
        <Case title={item.title} key={item.id} description={item.description} link={item.link} />
      ))}
    </div>
  );
};
