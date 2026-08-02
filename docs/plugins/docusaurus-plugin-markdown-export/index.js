// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

const fs = require('fs');
const path = require('path');
const {processMarkdownFile} = require('./mdxProcessor');

/**
 * Generates a clean .md file for every .mdx doc page at build time,
 * mirroring the Docusaurus URL structure:
 *
 *   content/getting-started/foo.mdx
 *     → build/docs/next/getting-started/foo.md
 *     → served at /docs/next/getting-started/foo.md
 *
 * ThunderID only has one docs version ("current" → served under /docs/next/).
 */
module.exports = function pluginMarkdownExport(context) {
  const {siteDir, siteConfig} = context;
  const siteUrl = (siteConfig?.url || '').replace(/\/$/, '');
  const baseUrl = siteConfig?.baseUrl || '/';

  // ThunderID keeps docs in content/ (not the Docusaurus default docs/)
  const DOCS_SOURCE_DIR = path.join(siteDir, 'content');
  // The "current" version is served under /docs/next/ in Docusaurus config
  const VERSION_URL_PREFIX = 'next/';

  /** Recursively collect all .md/.mdx files, skipping hidden/special ones. */
  function findMarkdownFiles(dir, baseDir = dir) {
    const files = [];
    if (!fs.existsSync(dir)) return files;

    for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        files.push(...findMarkdownFiles(fullPath, baseDir));
      } else if (/\.(md|mdx)$/.test(entry.name) && !entry.name.startsWith('_')) {
        const relativePath = path.relative(baseDir, fullPath);
        const slug = relativePath.replace(/\.(md|mdx)$/, '').split(path.sep).join('/');
        files.push({fullPath, slug});
      }
    }
    return files;
  }

  async function exportAll(outDir) {
    if (!fs.existsSync(DOCS_SOURCE_DIR)) {
      console.warn(`[markdown-export] Source dir not found: ${DOCS_SOURCE_DIR}`);
      return 0;
    }

    const files = findMarkdownFiles(DOCS_SOURCE_DIR);
    console.log(`[markdown-export] Processing ${files.length} docs`);
    let written = 0;

    for (const {fullPath, slug} of files) {
      // Output path mirrors the Docusaurus URL
      const outputPath = path.join(outDir, 'docs', VERSION_URL_PREFIX, slug + '.md');
      const docUrlPath = (baseUrl + 'docs/' + VERSION_URL_PREFIX + slug).replace(/\/+/g, '/');

      try {
        const source = fs.readFileSync(fullPath, 'utf-8');
        const cleaned = await processMarkdownFile(source, {}, path.dirname(fullPath), {
          docUrlPath,
          siteUrl,
        });

        fs.mkdirSync(path.dirname(outputPath), {recursive: true});
        fs.writeFileSync(outputPath, cleaned);
        written++;
      } catch (err) {
        console.error(`[markdown-export] Error processing ${fullPath}: ${err.message}`);
      }
    }

    return written;
  }

  return {
    name: 'docusaurus-plugin-markdown-export',

    async postBuild({outDir}) {
      const written = await exportAll(outDir);
      console.log(`[markdown-export] Wrote ${written} .md files to build/docs/${VERSION_URL_PREFIX}`);
    },
  };
};
