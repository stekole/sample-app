import http from 'k6/http';
import { check, sleep } from 'k6';

// Get base URL from environment variable with a fallback
const BASE_URL = __ENV.BASE_URL || 'https://example.com';

export const options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '15m', target: 1000 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // Keep performance threshold
    'checks{errorStatus:500}': ['rate>=0.99'], // For error scenarios
    'checks{status:200}': ['rate>=0.99'],      // For success scenarios
  },
  insecureSkipTLSVerify: true //important as we dont have a real cert

};

export function successTest() {
  const res = http.get(`${BASE_URL}/hello`);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response body': (r) => r.body.includes('Hello, World!'),
  });
  sleep(1);
}

export function errorTest() {
  const failRes = http.get(`${BASE_URL}/hello`, {
    headers: { 'Fail': 'true' },
  });
  check(failRes, {
    'error status is 500': (r) => r.status === 500,
  });
  sleep(1);
}

export default function () {  // Added default export
  successTest();
  errorTest();  // Uncomment to test error scenario
}
