import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: Number(__ENV.K6_VUS || 1),
  iterations: Number(__ENV.K6_ITERATIONS || 10),
  thresholds: {
    http_req_failed: ['rate==0'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ADMIN_USER = __ENV.AUTH_ADMIN_USERNAME || 'admin';
const ADMIN_PASS = __ENV.AUTH_ADMIN_PASSWORD || 'changeit';

function authenticate(scope) {
  // Omitting scope mints a token with all scopes granted to the user, the
  // same convention as internal/catalog/integration_test.go. oapi-codegen's
  // gin template collapses the spec's alternative security requirements
  // (admin OR manager) to the last one listed, so requesting a narrowed
  // 'admin' scope would fail writes that effectively require 'manager'.
  const body = { username: ADMIN_USER, password: ADMIN_PASS };
  if (scope) {
    body.scope = scope;
  }
  const res = http.post(`${BASE_URL}/auth/token`, JSON.stringify(body), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, {
    [`token ${scope} status 200`]: (r) => r.status === 200,
    [`token ${scope} has payload`]: (r) => !!(r.json('accessToken')),
  });
  return res.json('accessToken');
}

export default function () {
  const viewerToken = authenticate('viewer');
  const viewerRes = http.get(`${BASE_URL}/categories`, {
    headers: { Authorization: `Bearer ${viewerToken}` },
  });
  check(viewerRes, {
    'list categories success': (r) => r.status === 200,
  });

  const adminToken = authenticate();
  const createRes = http.post(
    `${BASE_URL}/categories`,
    JSON.stringify({ name: `k6 smoke ${Date.now()}` }),
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminToken}`,
      },
    },
  );
  check(createRes, {
    'create category created': (r) => r.status === 201,
  });

  const createdId = createRes.json('id');
  if (createdId) {
    const deleteRes = http.del(`${BASE_URL}/categories/${createdId}`, null, {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    check(deleteRes, {
      'delete category success': (r) => r.status === 204,
    });
  }

  sleep(1);
}
