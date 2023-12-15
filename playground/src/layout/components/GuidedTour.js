import React, { useState } from "react";
import { Tour } from "antd";

const Link = ({ href, text }) => (
  <a href={href} target="_blank" rel="noreferrer noopener">
    {text}
  </a>
);

export default function GuidedTour({ refSchemaEditor, refRelationshipsAndAttributes, refEnforcement }) {
  const [open, setOpen] = useState(!(localStorage.getItem("hasSeenGuidedTour") === "true"));

  const onTourClose = () => {
    localStorage.setItem("hasSeenGuidedTour", "true");
    setOpen(false);
  };

  const steps = [
    {
      title: "Welcome to the Permify Playground!",
      placement: "center",
      description: (
        <p>
          This environment enables you to create and test your authorization schema within a browser. The Permify playground comprises three main sections: Schema (Authorization Model), Authorization Data, and Enforcement. While we cover these
          sections in this tour, you can find the complete documentation at <Link href="https://docs.permify.co/" text="docs.permify.co" />.
        </p>
      ),
    },
    {
      title: "Schema (Authorization Model)",
      target: refSchemaEditor.current,
      description: (
        <p>
          You can create your authorization model in this section with using our domain specific language. We already have a couple of use cases and example that you can choose from the dropdown above. Also, you can check our{" "}
          <Link href="https://docs.permify.co/docs/getting-started/modeling/" text="docs" /> to learn more about how to model authorization in Permify.
        </p>
      ),
    },
    {
      title: "Authorization Data",
      target: refRelationshipsAndAttributes.current,
      description: (
        <p>
          You can create sample authorization data to test your authorization logic. For instance, to create a relationship between your defined entities, simply click the 'Add Relationship' button. For further information on data creation, please
          refer to <Link href="https://docs.permify.co/docs/getting-started/sync-data/" text="docs" />.
        </p>
      ),
    },
    {
      title: "Enforcement",
      target: refEnforcement.current,
      description: (
        <p>
          Now that we have sample data and a defined schema, let's perform an access check! The YAML in the Enforcement section represents a test scenario for conducting access checks. To learn more about the capabilities of this YAML, refer to:{" "}
          <Link href="https://docs.permify.co/docs/playground/#enforcement-access-check-scenarios" text="docs" />.
        </p>
      ),
    },
  ];
  return <Tour open={open} steps={steps} onClose={onTourClose} />;
}
