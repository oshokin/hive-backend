import http from 'k6/http';
import { check } from 'k6';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.3/index.js';
import { URL } from 'https://jslib.k6.io/url/1.0.0/index.js';

export function setup() {
    return {
        startDate: new Date()
    };
}

export default function () {
    const methodUrl = new URL('http://localhost:8080/v1/user/search');
    const params = {
        first_name: 'Олег',
        last_name: 'Шокин',
        limit: 1,
    };

    methodUrl.search = new URLSearchParams(params).toString();

    const res = http.get(methodUrl.href);

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