import { refreshStaleDefaultModels } from '../../src/models/defaultModelRefresher';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

describe('refreshStaleDefaultModels', () => {
  let tmpModelsDir: string;
  let tmpCanonicalDir: string;

  beforeEach(async () => {
    tmpModelsDir = await fs.mkdtemp(path.join(os.tmpdir(), 'models-'));
    tmpCanonicalDir = await fs.mkdtemp(path.join(os.tmpdir(), 'canonical-'));

    // Create canonical model files
    await fs.writeFile(path.join(tmpCanonicalDir, 'modelA.bin'), 'canonical content A', 'utf-8');
    await fs.writeFile(path.join(tmpCanonicalDir, 'modelB.bin'), 'canonical content B', 'utf-8');

    // Create model files and meta files in modelsDir
    // modelA is default and stale
    const modelAPath = path.join(tmpModelsDir, 'modelA.bin');
    await fs.writeFile(modelAPath, 'old content A', 'utf-8');
    const modelAMeta = {
      id: 'modelA',
      lastUpdated: Date.now() - 1000 * 60 * 60 * 24 * 10, // 10 days ago
      isDefault: true,
      filePath: modelAPath,
    };
    await fs.writeFile(modelAPath + '.meta.json', JSON.stringify(modelAMeta), 'utf-8');

    // modelB is default but fresh
    const modelBPath = path.join(tmpModelsDir, 'modelB.bin');
    await fs.writeFile(modelBPath, 'old content B', 'utf-8');
    const modelBMeta = {
      id: 'modelB',
      lastUpdated: Date.now(),
      isDefault: true,
      filePath: modelBPath,
    };
    await fs.writeFile(modelBPath + '.meta.json', JSON.stringify(modelBMeta), 'utf-8');

    // modelC is non-default and stale
    const modelCPath = path.join(tmpModelsDir, 'modelC.bin');
    await fs.writeFile(modelCPath, 'old content C', 'utf-8');
    const modelCMeta = {
      id: 'modelC',
      lastUpdated: Date.now() - 1000 * 60 * 60 * 24 * 10,
      isDefault: false,
      filePath: modelCPath,
    };
    await fs.writeFile(modelCPath + '.meta.json', JSON.stringify(modelCMeta), 'utf-8');
  });

  afterEach(async () => {
    // Cleanup temp dirs
    await fs.rm(tmpModelsDir, { recursive: true, force: true });
    await fs.rm(tmpCanonicalDir, { recursive: true, force: true });
  });

  it('refreshes only stale default models', async () => {
    const staleThresholdMs = 1000 * 60 * 60 * 24 * 7; // 7 days

    const refreshedCount = await refreshStaleDefaultModels({
      modelsDir: tmpModelsDir,
      staleThresholdMs,
      canonicalModelsDir: tmpCanonicalDir,
    });

    expect(refreshedCount).toBe(1);

    // Check modelA file content replaced with canonical content
    const modelAPath = path.join(tmpModelsDir, 'modelA.bin');
    const contentA = await fs.readFile(modelAPath, 'utf-8');
    expect(contentA).toBe('canonical content A');

    // Check modelA meta lastUpdated updated
    const metaRaw = await fs.readFile(modelAPath + '.meta.json', 'utf-8');
    const meta = JSON.parse(metaRaw);
    expect(meta.lastUpdated).toBeGreaterThan(Date.now() - 1000 * 60); // updated within last minute

    // modelB should be unchanged
    const modelBPath = path.join(tmpModelsDir, 'modelB.bin');
    const contentB = await fs.readFile(modelBPath, 'utf-8');
    expect(contentB).toBe('old content B');

    // modelC should be unchanged (not default)
    const modelCPath = path.join(tmpModelsDir, 'modelC.bin');
    const contentC = await fs.readFile(modelCPath, 'utf-8');
    expect(contentC).toBe('old content C');
  });

  it('returns 0 if no models are stale', async () => {
    const staleThresholdMs = 1000 * 60 * 60 * 24 * 20; // 20 days

    const refreshedCount = await refreshStaleDefaultModels({
      modelsDir: tmpModelsDir,
      staleThresholdMs,
      canonicalModelsDir: tmpCanonicalDir,
    });

    expect(refreshedCount).toBe(0);
  });

  it('handles malformed meta files gracefully', async () => {
    // Write malformed meta file
    await fs.writeFile(path.join(tmpModelsDir, 'bad.meta.json'), '{bad json', 'utf-8');

    const staleThresholdMs = 1000 * 60 * 60 * 24 * 7; // 7 days

    const refreshedCount = await refreshStaleDefaultModels({
      modelsDir: tmpModelsDir,
      staleThresholdMs,
      canonicalModelsDir: tmpCanonicalDir,
    });

    expect(refreshedCount).toBe(1);
  });
});
