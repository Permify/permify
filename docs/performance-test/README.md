# Permify Load Testing

For guideline purposes we perform load test on **Permify** with 1000 VU and 10000 RPS. This document contains our Permify schema, the grafana/k6 test script used for load testing, and documented results.

## Table of Contents
1. [Test Environment](#test-environment)
2. [Schema](#schema)
3. [K6 Test Script](#k6-test-script)
4. [Test Results](#test-results)
---

## 1. Test Environment

- Permify version **1.3.0**
- Google Kubernetes Engine general-purpose machines for clusters
- Postgres 15 on Google Cloud SQL 
- pgcat for server side pooling
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
Below is the k6test.js script used to measure load on Permify by performing both write and check requests:
```javascript
import http from 'k6/http';
import {check, sleep} from 'k6';

export let options = {
    cloud: {
        distribution: {
            distributionLabel1: {loadZone: 'amazon:de:frankfurt', percent: 100},
        },
    },
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

let reuseIdProbability = 0.3;
let currentId = getRandomId();

export default function () {
    let entityId, subjectId;

    // Decide whether to reuse the current ID for the entity
    if (Math.random() < reuseIdProbability) {
        entityId = currentId; // Reuse the existing ID
    } else {
        entityId = getRandomId(); // Generate a new ID
        currentId = entityId; // Update current ID to the new one
    }

    // Decide whether to reuse the current ID for the subject
    if (Math.random() < reuseIdProbability) {
        subjectId = currentId; // Reuse the existing ID
    } else {
        subjectId = getRandomId(); // Generate a new ID
        currentId = subjectId; // Update current ID to the new one
    }

    const url = '<your-api-endpoint>';
    const payload = JSON.stringify({
        metadata: {
            snap_token: "",
            schema_version: "<schema-version>",
            depth: 20
        },
        entity: {
            type: "content",
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
| **checks**                     | 100.00% (75369 out of 75369)                                                  |
| **data_received**              | 15 MB (145 kB/s)                                                              |
| **data_sent**                  | 21 MB (203 kB/s)                                                              |
| **dropped_iterations**         | 271664 (2688.447952/s)                                                        |
| **http_req_blocked**           | avg=78.61µs min=208ns med=537ns max=45.53ms p(90)=775ns p(95)=885ns           |
| **http_req_connecting**        | avg=16.27µs min=0s med=0s max=15.27ms p(90)=0s p(95)=0s                       |
| **http_req_duration**          | avg=10.14ms min=1.61ms med=7.44ms max=295.29ms p(90)=14.3ms p(95)=26.34ms     |
| **{ expected_response:true }** | avg=10.14ms min=1.61ms med=7.44ms max=295.29ms p(90)=14.3ms p(95)=26.34ms     |
| **http_req_failed**            | 0.00% (0 out of 75369)                                                        |
| **http_req_receiving**         | avg=94.63µs min=15.55µs med=70.42µs max=37.13ms p(90)=168.34µs p(95)=234.35µs |
| **http_req_sending**           | avg=86.07µs min=25.5µs med=69.44µs max=10.25ms p(90)=121.44µs p(95)=171.32µs  |
| **http_req_tls_handshaking**   | avg=59.38µs min=0s med=0s max=44.21ms p(90)=0s p(95)=0s                       |
| **http_req_waiting**           | avg=9.96ms min=1.41ms med=7.27ms max=295.21ms p(90)=14.1ms p(95)=26.17ms      |
| **http_reqs**                  | 75369 (745.86855/s)                                                           |
| **iteration_duration**         | avg=1.01s min=1s med=1s max=1.29s p(90)=1.01s p(95)=1.02s                     |
| **iterations**                 | 75369 (745.86855/s)                                                           |
| **vus**                        | 46 (min=13, max=1000)                                                         |
| **vus_max**                    | 1000 (min=50, max=1000)                                                       |