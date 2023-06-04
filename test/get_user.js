import http from 'k6/http';
import { check } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.3/index.js';

export function setup() {
    return {
        startDate: new Date()
    };
}

export default function () {
    const userId = randomIntBetween(1, 1000000);
    const res = http.get(`http://localhost:8080/v1/user/${userId}`,
        {
            tags: { name: 'http://localhost:8080/v1/user/{id}' },
        });

    check(res, {
        'status was 200': (r) => r.status === 200,
    });
}

export function handleSummary(data) {
    const customizedReport = Object.assign(
        {
            startDate: new Date(data.setup_data.startDate),
            endDate: new Date(),
        },
        data
    );

    const summary = textSummary(customizedReport,
        {
            indent: ' ',
            enableColors: true,
        }
    );

    return {
        'stdout': `${summary}\n\nstarted: ${new Date(customizedReport.startDate)}\nended: ${customizedReport.endDate}\n`,
    };
}