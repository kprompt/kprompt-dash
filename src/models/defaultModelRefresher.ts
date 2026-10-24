/**
 * Module to refresh stale default models.
 * 
 * This module exports a function `refreshStaleDefaultModels` that checks the age of default models,
 * and refreshes them if they are stale beyond a configured threshold.
 * 
 * The refresh operation replaces the stale model with a fresh copy from the canonical source.
 */

import fs from 'fs/promises';
import path from 'path';

export interface ModelMeta {
  id: string;
  lastUpdated: number; // timestamp in ms
  isDefault: boolean;
  filePath: string;
}

export interface RefreshOptions {
  modelsDir: string;
  staleThresholdMs: number;
  canonicalModelsDir: string;
}

/**
 * Loads metadata for all models in the models directory.
 * Assumes metadata is stored in a JSON file alongside the model file, named `${id}.meta.json`.
 */
async function loadModelsMeta(modelsDir: string): Promise<ModelMeta[]> {
  const files = await fs.readdir(modelsDir);
  const metaFiles = files.filter(f => f.endsWith('.meta.json'));
  const metas: ModelMeta[] = [];
  for (const metaFile of metaFiles) {
    try {
      const metaPath = path.join(modelsDir, metaFile);
      const metaRaw = await fs.readFile(metaPath, 'utf-8');
      const meta = JSON.parse(metaRaw);
      if (
        typeof meta.id === 'string' &&
        typeof meta.lastUpdated === 'number' &&
        typeof meta.isDefault === 'boolean' &&
        typeof meta.filePath === 'string'
      ) {
        metas.push(meta);
      }
    } catch {
      // ignore malformed meta files
    }
  }
  return metas;
}

/**
 * Refreshes a single model by copying the canonical model file over the existing one,
 * and updating the lastUpdated timestamp in the meta file.
 */
async function refreshModel(model: ModelMeta, canonicalModelsDir: string): Promise<void> {
  const canonicalModelPath = path.join(canonicalModelsDir, path.basename(model.filePath));
  const modelFilePath = model.filePath;
  const metaFilePath = modelFilePath + '.meta.json';

  // Copy canonical model file over
  await fs.copyFile(canonicalModelPath, modelFilePath);

  // Update meta file lastUpdated timestamp
  const newMeta = { ...model, lastUpdated: Date.now() };
  await fs.writeFile(metaFilePath, JSON.stringify(newMeta, null, 2), 'utf-8');
}

/**
 * Main function to refresh stale default models.
 * 
 * @param opts RefreshOptions
 * @returns number of models refreshed
 */
export async function refreshStaleDefaultModels(opts: RefreshOptions): Promise<number> {
  const { modelsDir, staleThresholdMs, canonicalModelsDir } = opts;
  const metas = await loadModelsMeta(modelsDir);
  const now = Date.now();

  let refreshedCount = 0;
  for (const model of metas) {
    if (!model.isDefault) continue;
    if (now - model.lastUpdated > staleThresholdMs) {
      try {
        await refreshModel(model, canonicalModelsDir);
        refreshedCount++;
      } catch (err) {
        // Log and continue, fail closed by not refreshing this model
        console.error(`Failed to refresh model ${model.id}:`, err);
      }
    }
  }
  return refreshedCount;
}
