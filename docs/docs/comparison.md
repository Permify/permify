---
id: comparison
title: Comparison Between Other Zanzibar implementations
---

:::caution Note
This comparison table shows the differentiation between authorization solutions based or inspired by Google Zanzibar paper. If you use any of these solutions and feel the information could be improved, feel free to reach out.
:::

## General Aspects

|                                 | Ory/Keto   | OpenFGA    | SpiceDB   | Permify   |
|---------------------------------|------------|------------|-----------|-----------|
| **Zanzibar Paper Faithfulness** | Medium     | High       | High      | High      |
| **Scalability**                 | Medium     | Medium     | High      | High      |
| **Consistency & Cache**         | No Zookies | No Zookies | Supported | Supported |
| **Dev UX**                      | Average    | Average    | High      | High      |

## Feature Set

-   ✅ &nbsp;Supported, and ready to use with no added configuration or code
-   🟡 &nbsp;Limited support and requires extra user-code to implement.
-   ⛔ &nbsp;Not officially supported or documented.

|                          | Ory/Keto | OpenFGA | SpiceDB | Permify |
|--------------------------|----------|---------|---------|---------|
| **Check API**            | ✅        | ✅       | ✅       | ✅       |
| **Write API**            | ✅        | ✅       | ✅       | ✅       |
| **Read API**             | ✅        | ✅       | ✅       | ✅       |
| **Expand API**           | ✅        | ✅       | ✅       | ✅       |
| **Watch API**            | ✅        | ✅       | ✅       | ✅       |
| **RBAC**                 | ✅        | ✅       | ✅       | ✅       |
| **ReBAC**                | ✅        | ✅       | ✅       | ✅       |
| **ABAC**                 | ⛔        | 🟡      | ✅       | ✅       |
| **Data Filtering**       | ⛔        | ✅       | ✅       | ✅       |
| **Multi Tenancy**        | ⛔        | ✅       | ⛔       | ✅       |
| **Testing & Validation** | ⛔        | 🟡      | ✅       | ✅       |
| **Logging & Tracing**    | 🟡       | ✅       | ✅       | ✅       |
