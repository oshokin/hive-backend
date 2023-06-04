import sql from 'k6/x/sql';
import { check } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'
import { Counter } from 'k6/metrics';

const successfulTxCount = new Counter('successful_tx_count');
const failedTxCount = new Counter('failed_tx_count');
const uniqueErrors = new Set();

const db = sql.open('postgres', 'postgres://admin:hard-password@localhost:5432/hive?sslmode=disable');

export function setup() {
    try {
        db.exec(`
        CREATE TABLE IF NOT EXISTS random_data (
            id BIGSERIAL PRIMARY KEY,
            random_number INTEGER NOT NULL
        );
        `);
    } catch (error) {
        logUniqueError('failed to run setup phase: ' + error)
    }
}

export default function () {
    const randomNumber = randomIntBetween(1, 1000000);
    let result;
    try {
        result = db.exec(`INSERT INTO random_data (random_number) VALUES (${randomNumber});`);
    } catch (error) {
        logUniqueError('failed to run test: ' + error)
        failedTxCount.add(1);

        check(result, {
            'data stored successfully': false,
        });

        return
    }

    if (!result || result.error) {
        failedTxCount.add(1);
    } else {
        successfulTxCount.add(1);
    }

    check(result, {
        'data stored successfully': (r) => !r.error,
    });
}

export function teardown() {
    try {
        db.exec('DROP TABLE IF EXISTS random_data;');
        db.close();
    } catch (error) {
        logUniqueError('failed to run teardown phase: ' + error)
    }
}

function logUniqueError(error) {
    if (uniqueErrors.has(error)) {
        return
    }

    uniqueErrors.add(error);
    console.error(error);
}