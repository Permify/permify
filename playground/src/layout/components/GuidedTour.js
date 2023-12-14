import React, { useState } from "react";
import { Tour } from "antd";

export default function GuidedTour({ refSchemaEditor, refRelationshipsAndAttributes, refEnforcement }) {
  const [open, setOpen] = useState(true);
  const steps = [
    {
      title: "Welcome to the Permify Playground!",
      placement: "center",
      description: (
        <p>
          This environment enables you to create and test your authorization schema within a browser. The Permify playground comprises three main sections: Schema (Authorization Model), Authorization Data, and Enforcement. While we cover these
          sections in this tour, you can find the complete documentation at &nbsp;
          <a href="https://docs.permify.co/" target="_blank" rel="noreferrer noopener">
            docs.permify.co
          </a>
        </p>
      ),
    },
    {
      title: "Schema (Authorization Model)",
      target: refSchemaEditor.current,
      description: (
        <p>
          You can create your authorization model in this section with using our domain specific language. We already have a couple of use cases and example that you can choose from the dropdown above. Also, you can check our &nbsp;
          <a href="https://docs.permify.co/docs/getting-started/modeling/" target="_blank" rel="noreferrer noopener">
            docs
          </a>
          &nbsp; to learn more about how to model authorization in Permify.
        </p>
      ),
    },
    {
      title: "Authorization Data",
      target: refRelationshipsAndAttributes.current,
      description: (
        <p>
          You can create sample authorization data to test your authorization logic. For instance, to create a relationship between your defined entities, simply click the 'Add Relationship' button. For further information on data creation, please
          refer to &nbsp;
          <a href="https://docs.permify.co/docs/getting-started/sync-data/" target="_blank" rel="noreferrer noopener">
            docs
          </a>
        </p>
      ),
    },
    {
      title: "Enforcement",
      target: refEnforcement.current,
      description: (
        <p>
          Now that we have sample data and a defined schema, let's perform an access check! The YAML in the Enforcement section represents a test scenario for conducting access checks. To learn more about the capabilities of this YAML, refer to:
          &nbsp;
          <a href="https://docs.permify.co/docs/playground/#enforcement-access-check-scenarios" target="_blank" rel="noreferrer noopener">
            access check scenarios
          </a>
        </p>
      ),
    },
  ];
  return <Tour open={open} steps={steps} onClose={() => setOpen(false)} />;
}
