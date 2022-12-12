# Delete Relational Tuples

You can delete any stored relation tuples with following path

**Path:** POST /v1/relationships/delete

| Required | Argument | Type | Default | Description |
|----------|----------|---------|---------|-------------------------------------------------------------------------------------------|
| [x]   | entity | object | - | contains entity type and id of the entity. Example: repository:1”.
| [x]   | relation | string | - | relation of the given entity |
| [ ]   | subject | object | - | the user or user set. It containes type and id of the subject.  ||

#### Request

```json
{
  "filter": {
    "entity": {
      "type": "organization",
      "ids": [
        "1"
      ]
    },
    "relation": "admin",
    "subject": {
      "type": "user",
      "ids": [
        "1"
      ],
      "relation": ""
    }
  }
}
```

## Need any help ?

Our team is happy to help you get started with Permify. If you'd like to learn more about using Permify in your app or have any questions about this example, [schedule a call with one of our Permify engineer](https://meetings-eu1.hubspot.com/ege-aytin/call-with-an-expert).