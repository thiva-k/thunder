// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

class NodeSqliteDatabase {
  constructor(DatabaseSync, path) {
    this.database = new DatabaseSync(path);
  }

  close() {
    return this.database.close();
  }

  exec(sql) {
    return this.database.exec(sql);
  }

  prepare(sql) {
    return this.database.prepare(sql);
  }

  transaction(callback) {
    return (...args) => {
      this.exec("BEGIN TRANSACTION");

      try {
        const result = callback(...args);
        this.exec("COMMIT");

        return result;
      } catch (error) {
        this.exec("ROLLBACK");

        throw error;
      }
    };
  }
}

export function isNodeSqlitePublicVersion(version = process.versions.node) {
  const [major = 0, minor = 0] = version.split(".").map(Number);

  return major > 23 || (major === 23 && minor >= 4) || (major === 22 && minor >= 13);
}

async function loadDatabaseSync() {
  const provider = process.env.WAYFINDER_SQLITE_PROVIDER || "auto";

  if (provider !== "auto" && provider !== "node:sqlite" && provider !== "better-sqlite3") {
    throw new Error(`Unsupported WAYFINDER_SQLITE_PROVIDER: ${provider}`);
  }

  if (provider === "node:sqlite" || (provider === "auto" && isNodeSqlitePublicVersion())) {
    try {
      const { DatabaseSync } = await import("node:sqlite");

      return class Database extends NodeSqliteDatabase {
        constructor(path) {
          super(DatabaseSync, path);
        }
      };
    } catch (error) {
      if (provider === "node:sqlite" || error.code !== "ERR_UNKNOWN_BUILTIN_MODULE") {
        throw error;
      }
    }
  }

  try {
    const { default: Database } = await import("better-sqlite3");

    return Database;
  } catch (error) {
    if (provider === "better-sqlite3") {
      throw error;
    }

    throw new Error(
      "SQLite requires Node.js with node:sqlite support or an installed better-sqlite3 fallback.",
      { cause: error },
    );
  }
}

export const DatabaseSync = await loadDatabaseSync();
