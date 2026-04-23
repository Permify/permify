# Permify Load Testing

For guideline purposes we perform load test on **Permify** with 1000 VU and 10000 RPS. This document contains our Permify schema, the grafana/k6 test script used for load testing, and documented results.

## Table of Contents
1. [Test Environment](#test-environment)
2. [Schema](#schema)
3. [K6 Test Script](#k6-test-script)
4. [Test Results](#test-results)
---

## 1. Test Environment

- Permify version **1.6.9**
- Google Kubernetes Engine general-purpose machines for clusters
- Postgres 15 on Google Cloud SQL 
- **[values.yaml](./values.yaml)**: This file holds the default configuration (e.g. Helm values, resource settings, etc.) used for the performance test environment.

## 2. Schema

Below is the schema we use for our load tests:

```
entity user {
    relation self @user
    relation follower @user
    relation blocked @user
    
    attribute is_public boolean

    permission view = self or (is_public or follower) not blocked
}

entity content {
    relation owner @user
    attribute is_public boolean

    permission view = owner.self or (is_public or owner.follower) not owner.blocked
}

entity interaction {
    relation creator @user
    relation parent @content

    permission view = creator.self or (creator.view and parent.view)
}
```

## 3. K6 Test Script

### Write Script
Seeds data for the schema used in our load tests.
```javascript
import http from "k6/http";
import { fail } from "k6";

export const options = { vus: 1, iterations: 1 };

const TENANT_ID = "<tenant-id>";
const url = "<your-api-endpoint>";

const TOTAL_IDS = 100000;
const BATCH_SIZE = 1000;

function writeInBatches(items, key) {
    for (let start = 0; start < items.length; start += BATCH_SIZE) {
        const chunk = items.slice(start, start + BATCH_SIZE);
        const batchNumber = start / BATCH_SIZE + 1;
        const res = http.post(
            `${url}/v1/tenants/${TENANT_ID}/data/write`,
            JSON.stringify({
                metadata: { schema_version: "<schema-version>" },
                [key]: chunk,
            }),
            {
                headers: { "Content-Type": "application/json" },
            }
        );

        if (res.status < 200 || res.status >= 300) {
            console.error(`\nPOST /data/write ${key} batch ${batchNumber} FAILED status=${res.status}\nbody=\n${res.body}\n`);
            fail(`POST /data/write ${key} batch ${batchNumber} non-2xx`);
        }
    }
}

export default function () {
    const tuples = [];
    const attributes = [];

    for (let i = 0; i < TOTAL_IDS; i++) {
        const id = String(i);
        const nextId = String((i + 1) % TOTAL_IDS);
        const blockedId = String((i + 2) % TOTAL_IDS);
        const isPublic = i % 4 === 0;

        tuples.push(
            {
                entity: { type: "user", id },
                relation: "self",
                subject: { type: "user", id, relation: "" },
            },
            {
                entity: { type: "user", id },
                relation: "follower",
                subject: { type: "user", id: nextId, relation: "" },
            },
            {
                entity: { type: "user", id },
                relation: "blocked",
                subject: { type: "user", id: blockedId, relation: "" },
            },
            {
                entity: { type: "content", id },
                relation: "owner",
                subject: { type: "user", id, relation: "" },
            },
            {
                entity: { type: "interaction", id },
                relation: "creator",
                subject: { type: "user", id, relation: "" },
            },
            {
                entity: { type: "interaction", id },
                relation: "parent",
                subject: { type: "content", id, relation: "" },
            }
        );

        attributes.push(
            {
                entity: { type: "user", id },
                attribute: "is_public",
                value: { boolean: isPublic },
            },
            {
                entity: { type: "content", id },
                attribute: "is_public",
                value: { boolean: isPublic },
            }
        );
    }

    writeInBatches(tuples, "tuples");
    writeInBatches(attributes, "attributes");

    console.log(`Seeded ${TOTAL_IDS} users, ${TOTAL_IDS} contents, and ${TOTAL_IDS} interactions`);
}
```

### Test Script

Below is the k6test.js script used to measure load on Permify by performing check requests:
```javascript
import http from 'k6/http';
import {check, sleep} from 'k6';

export let options = {
    scenarios: {
        contacts_load: {
            executor: 'ramping-arrival-rate',
            startRate: 10,  // starting rate of new iterations per timeUnit
            timeUnit: '1s', // new iterations per second
            preAllocatedVUs: 50, // minimum number of VUs before the test starts
            maxVUs: 100,    // maximum number of VUs during the test
            stages: [
                {target: 100, duration: '10s'},  // Warm-up phase
                {target: 1000, duration: '30s'},  // Warm-up phase
                {target: 10000, duration: '1m' }, // Ramp up to full load
            ]
        }
    }
};

function getRandomId() {
    return Math.floor(Math.random() * 100000).toString();
}

let reuseIdProbability = 0.1;
let currentEntityId = getRandomId();
let currentSubjectId = getRandomId();

export default function () {
    let entityId, subjectId;
    const entityType = Math.random() < 0.5 ? 'content' : 'interaction';

    // Decide whether to reuse the current ID for the entity
    if (Math.random() < reuseIdProbability) {
        entityId = currentEntityId; // Reuse the existing entity ID
    } else {
        entityId = getRandomId(); // Generate a new entity ID
        currentEntityId = entityId; // Update current entity ID
    }

    // Decide whether to reuse the current ID for the subject
    if (Math.random() < reuseIdProbability) {
        subjectId = currentSubjectId; // Reuse the existing subject ID
    } else {
        subjectId = getRandomId(); // Generate a new subject ID
        currentSubjectId = subjectId; // Update current subject ID
    }

    const url = '<your-api-endpoint>';
    const payload = JSON.stringify({
        metadata: {
            snap_token: "",
            schema_version: "<schema-version>",
            depth: 20
        },
        entity: {
            type: entityType,
            id: entityId,
        },
        permission: 'view',
        subject: {
            type: 'user',
            id: subjectId
        },
        page_size: 20
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer <your-token>`
        },
        timeout: 360000
    };

    let response = http.post(url, payload, params);
    //console.log(response);
    check(response, {
        "is status 200": (r) => r.status === 200
    });
    sleep(1);
}
```
## 3. How to Run
**Permify**

-  Running locally on http://localhost:3476 (default).

-  Update the api endpoint in k6test.js.

**k6**

-  [Install k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) using your preferred method (e.g., Homebrew, Chocolatey, Docker).

**Schema & Version**

- Write the schema ([Schema](#schema)).

- Set the schema version in `k6test.js` by updating the `SCHEMA_VERSION` variable.


## 4. Test Results

### Read Test Results

| **Metric**                     | **Value/Stats**                                                               |
|--------------------------------|-------------------------------------------------------------------------------|
| **checks**                     | 100.00% (74614 out of 74614)                                                  |
| **data_received**              | 17 MB (168 kB/s)                                                              |
| **data_sent**                  | 27 MB (268 kB/s)                                                              |
| **dropped_iterations**         | 272433 (2696.482348/s)                                                        |
| **http_req_blocked**           | avg=5.59µs  min=0s    med=2µs     max=2.08ms   p(90)=4µs     p(95)=5µs        |
| **http_req_connecting**        | avg=2.57µs  min=0s    med=0s      max=2.04ms   p(90)=0s      p(95=0s          |
| **http_req_duration**          | avg=21.3ms  min=428µs med=15.38ms max=617.85ms p(90)=45.7ms  p(95)=58.99ms    |
| **expected_response**          | avg=21.3ms  min=428µs med=15.38ms max=617.85ms p(90)=45.7ms  p(95)=58.99ms    |
| **http_req_failed**            | 0.00% (0 out of 74614)                                                        |
| **http_req_receiving**         | avg=20.27µs min=4µs   med=17µs    max=2.86ms   p(90)=37µs    p(95)=44µs       |
| **http_req_sending**           | avg=11.16µs min=1µs   med=8µs     max=2.51ms   p(90)=18µs    p(95)=22µs       |
| **http_req_tls_handshaking**   | avg=59.38µs min=0s med=0s max=44.21ms p(90)=0s p(95)=0s                       |
| **http_req_waiting**           | avg=21.27ms min=399µs med=15.35ms max=617.83ms p(90)=45.67ms p(95)=58.96ms    |
| **http_reqs**                  | 74614 (738.51308/s)                                                           |
| **iteration_duration**         | avg=1.02s   min=1s    med=1.01s   max=1.61s    p(90)=1.04s   p(95)=1.05s      |
| **iterations**                 | 74614 (738.51308/s)                                                           |
| **vus**                        | 114 (min=14, max=1000)                                                        |
| **vus_max**                    | 1000 (min=50, max=1000)                                                       |