#!/usr/bin/env node

const fs = require("fs")
const path = require("path")

const rawArg = process.argv[2]

if (!rawArg || rawArg === "-h" || rawArg === "--help") {
  console.error("Usage: ./scripts/bump.js <version>")
  process.exit(1)
}

const normalized = rawArg.startsWith("v") ? rawArg.slice(1) : rawArg
const isSemver = /^\d+\.\d+\.\d+$/.test(normalized)

if (!isSemver) {
  console.error(`Invalid version "${rawArg}". Expected x.y.z or vx.y.z.`)
  process.exit(1)
}

const configPath = path.resolve(__dirname, "..", "build", "config.yml")
const raw = fs.readFileSync(configPath, "utf8")
const lines = raw.split(/\r?\n/)

let inInfo = false
let updated = false
let sawInfo = false

for (let i = 0; i < lines.length; i += 1) {
  const line = lines[i]

  if (!inInfo) {
    if (/^info:\s*$/.test(line)) {
      inInfo = true
      sawInfo = true
    }
    continue
  }

  if (/^\S/.test(line)) {
    inInfo = false
  }

  if (!inInfo || updated) {
    continue
  }

  const match = line.match(/^(\s{2}version:\s*)(["']?)([^"']*)(["']?)(\s*#.*)?$/)
  if (match) {
    const comment = match[5] ?? ""
    lines[i] = `${match[1]}"${normalized}"${comment}`
    updated = true
  }
}

if (!sawInfo) {
  console.error("Failed to locate info block in build/config.yml.")
  process.exit(1)
}

if (!updated) {
  console.error("Failed to locate info.version in build/config.yml.")
  process.exit(1)
}

const next = lines.join("\n")
const output = raw.endsWith("\n") ? next : `${next}\n`
fs.writeFileSync(configPath, output)

console.log(`Updated build/config.yml info.version to ${normalized}`)
