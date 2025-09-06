#!/usr/bin/env node

const fs = require('fs');
const os = require('os');
const path = require('path');
const https = require('https');
const { spawn } = require('child_process');
const { createHash } = require('crypto');

const artifactoryUrl = 'https://artifacts.na.c-gurus.com/artifactory';
const binaryFile = path.join(cacheDir(), `usi${bazelPlatform()}`);
const binaryUrl = `${artifactoryUrl}/usi/${bazelPlatform()}`;
const shaFile = `${binaryFile}.sha256`;
const shaUrl = `${artifactoryUrl}/usi/${bazelPlatform()}`;
const sleepSeconds = 10;
const maxRetries = 10;

const isQuiet = process.argv.includes("-q") // If the cli arquments contain `-q` then it'll silence update logs

class CustomLogger {
    /**
     * @param {{quiet: boolean}} options
     */
    constructor({quiet}) {
        this.silenced = quiet
    }
    log(...args) { if (!this.silenced) console.log(...args); }
    error(...args) { if (!this.silenced) console.error(...args); }
}

const logger = new CustomLogger({quiet:isQuiet})

async function main() {
    await ensureLatest();
    // Establish pipes between parent and child for stdin, stdout, stderr
    // (stdin is needed to work nicely on the cmdline with tools like 'less')
    const cp = spawn(binaryFile, process.argv.slice(2), { stdio: 'inherit' });
    cp.on('close', code => process.exit(code));
}

main().catch(err => {
    console.error(err);
    process.exit(1);
});

async function ensureLatest() {
    let i = 0;
    do {
        [currentChecksum, err] = await isLatest();
        if (currentChecksum == "" && err == "") {
            return
        }

        if (err == "") {
            if (i == 0) {
                logger.log("updating usi cli...")
            }
            const content = await get(binaryUrl);
            fs.writeFileSync(binaryFile, content, { mode: 0o777, flags: 'w' });
            const latestChecksum = hash(content);
            if (latestChecksum === currentChecksum) {
                logger.log("usi updated")
                fs.writeFileSync(shaFile, hash(content));
                if (isDarwin()) {
                    const cp = spawn("/usr/bin/codesign", ['-s', '-', binaryFile], { stdio: 'inherit' });
                }
                return;
            } else {
                logger.error(`iteration ${i}: usi failed checksum validation: ${currentChecksum} vs ${latestChecksum}. Retrying...`)
            }
        } else {
            logger.error(`iteration ${i}: ${err}. Retrying...`)
        }
        i++
        if (i < maxRetries) {
            await sleep(sleepSeconds)
        }
    } while (i < maxRetries);

    logger.error(`Failed to update usi after ${maxRetries} attempts. Continuing...`)
}

function hash(content) {
    return createHash("sha256").update(content).digest("hex");
}

async function isLatest() {
    try {
        const response = await get(shaUrl);
        const latestSha256 = JSON.parse(response.toString()).checksums.sha256
        if (!fs.existsSync(binaryFile) || !fs.existsSync(shaFile) || latestSha256 !== fs.readFileSync(shaFile).toString()) {
            return [latestSha256, ""];
        } else {
            return ["", ""]
        }
    } catch (err) {
        return ["", "Error retrieving latest usi cli digest"]
    }
}

function isDarwin() {
    return os.platform() === 'darwin'
}

function bazelPlatform() {
    switch (os.platform()) {
        case 'darwin':
            if (process.arch === 'arm64') {
                return 'darwin_arm64';
            }
            return 'darwin_amd64';
        default:
            return 'linux_amd64';
    }
}

/**
 * Determine where we can cache files (USI binary and checksum).
 *
 * Follows the same logic as https://pkg.go.dev/os#UserCacheDir
 */
function cacheDir() {
    let cacheDir = process.env.XDG_CACHE_HOME;
    if (!cacheDir || !fs.existsSync(cacheDir)) {
        if (os.platform() === 'darwin') {
            cacheDir = path.join(process.env.HOME, 'Library', 'Caches');
        } else {
            cacheDir = path.join(process.env.HOME, '.cache');
        }
    }
    cacheDir = path.join(cacheDir, 'USIWrapper');
    if (!fs.existsSync(cacheDir)) {
        fs.mkdirSync(cacheDir, { recursive: true });
    }
    return cacheDir;
}

async function get(url) {
    return new Promise((resolve, reject) => {
        https.get(url, resp => {
            let data = [];
            resp.on('data', chunk => {
                data.push(chunk);
            }).on('end', () => {
                resolve(Buffer.concat(data));
            });
        }).on("error", (err) => {
            reject(err);
        });
    });
}

const sleep = seconds => new Promise(r => setTimeout(r, seconds * 1000));