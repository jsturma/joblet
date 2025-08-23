import express from 'express';
import {execRnx} from '../utils/rnxExecutor.js';

const router = express.Router();

// Get detailed system info
router.get('/system-info', async (req, res) => {
    try {
        const node = req.query.node;

        // Get comprehensive system information using monitor status command
        const output = await execRnx(['monitor', 'status', '--json'], {
            node,
            maxBuffer: 1024 * 1024 * 10 // 10MB buffer
        });

        const monitorData = JSON.parse(output);

        // Transform the data to match the frontend DetailedSystemInfo interface
        const systemInfo = {
            hostInfo: {
                hostname: monitorData.host?.hostname,
                platform: monitorData.host?.os,
                arch: monitorData.host?.architecture,
                release: monitorData.host?.kernelVersion,
                uptime: monitorData.host?.uptime,
                cloudProvider: monitorData.cloud?.provider,
                instanceType: monitorData.cloud?.instanceType,
                region: monitorData.cloud?.region
            },
            cpuInfo: {
                cores: monitorData.cpu?.cores,
                model: 'N/A', // Not provided by monitor status
                frequency: null, // Not provided by monitor status
                usage: monitorData.cpu?.usagePercent,
                loadAverage: monitorData.cpu?.loadAverage,
                perCoreUsage: monitorData.cpu?.perCoreUsage,
                temperature: null // Not provided by monitor status
            },
            memoryInfo: {
                total: monitorData.memory?.totalBytes,
                used: monitorData.memory?.usedBytes,
                available: monitorData.memory?.availableBytes,
                percent: monitorData.memory?.usagePercent,
                buffers: monitorData.memory?.bufferedBytes,
                cached: monitorData.memory?.cachedBytes,
                swap: {
                    total: monitorData.swap?.totalBytes || 0,
                    used: monitorData.swap?.usedBytes || 0,
                    percent: monitorData.swap?.usagePercent || 0
                }
            },
            diskInfo: monitorData.disks?.map(disk => ({
                device: disk.device,
                mountpoint: disk.mountPoint,
                filesystem: disk.filesystem,
                size: disk.totalBytes,
                used: disk.usedBytes,
                available: disk.freeBytes,
                percent: disk.usagePercent
            })) || [],
            networkInfo: [], // Not provided by monitor status
            processInfo: {
                totalProcesses: monitorData.processes?.totalProcesses || 0,
                totalThreads: monitorData.processes?.totalThreads || 0,
                runningProcesses: 0, // Not provided by monitor status
                sleepingProcesses: 0 // Not provided by monitor status
            }
        };

        res.json(systemInfo);
    } catch (error) {
        console.error('Failed to get system info:', error);
        res.status(500).json({error: error.message});
    }
});

// Get system metrics
router.get('/system-metrics', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['monitor', 'status', '--json'], {node});
        const metrics = JSON.parse(output);
        res.json(metrics);
    } catch (error) {
        console.error('Failed to get system metrics:', error);
        res.status(500).json({error: error.message});
    }
});

// Get volumes
router.get('/volumes', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['volume', 'list', '--json'], {node});

        let result;
        if (output && output.trim()) {
            try {
                result = JSON.parse(output);
                // Ensure volumes field exists and is an array
                if (!result.volumes || !Array.isArray(result.volumes)) {
                    result = {volumes: [], message: 'No volumes found'};
                }
            } catch (e) {
                console.warn('Failed to parse JSON from volume list:', e.message);
                result = {volumes: [], message: 'Volume service not available'};
            }
        } else {
            result = {volumes: [], message: 'Volume service not available'};
        }

        res.json(result);
    } catch (error) {
        console.error('Failed to list volumes:', error);
        res.json({volumes: [], message: `Volume service not available: ${error.message}`});
    }
});

// Create volume
router.post('/volumes', async (req, res) => {
    try {
        const {name, size, type} = req.body;
        const node = req.query.node;
        
        if (!name) {
            return res.status(400).json({error: 'Volume name is required'});
        }
        
        const args = ['volume', 'create', name];
        if (size) args.push('--size', size);
        if (type) args.push('--type', type);
        
        const output = await execRnx(args, {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to create volume:', error);
        res.status(500).json({error: error.message});
    }
});

// Delete volume
router.delete('/volumes/:volumeName', async (req, res) => {
    try {
        const {volumeName} = req.params;
        const node = req.query.node;
        
        const output = await execRnx(['volume', 'delete', volumeName], {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to delete volume:', error);
        res.status(500).json({error: error.message});
    }
});

// Get networks
router.get('/networks', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['network', 'list', '--json'], {node});

        let result;
        if (output && output.trim()) {
            try {
                result = JSON.parse(output);
                // Ensure networks field exists and is an array
                if (!result.networks || !Array.isArray(result.networks)) {
                    result = {
                        networks: [
                            {id: 'bridge', name: 'bridge', type: 'bridge', subnet: '172.17.0.0/16'},
                            {id: 'host', name: 'host', type: 'host', subnet: ''}
                        ],
                        message: 'Using default networks'
                    };
                }
            } catch (e) {
                console.warn('Failed to parse JSON from network list:', e.message);
                result = {
                    networks: [
                        {id: 'bridge', name: 'bridge', type: 'bridge', subnet: '172.17.0.0/16'},
                        {id: 'host', name: 'host', type: 'host', subnet: ''}
                    ],
                    message: 'Network service not available, showing defaults'
                };
            }
        } else {
            result = {
                networks: [
                    {id: 'bridge', name: 'bridge', type: 'bridge', subnet: '172.17.0.0/16'},
                    {id: 'host', name: 'host', type: 'host', subnet: ''}
                ],
                message: 'Network service not available, showing defaults'
            };
        }

        res.json(result);
    } catch (error) {
        console.error('Failed to list networks:', error);
        res.json({
            networks: [
                {id: 'bridge', name: 'bridge', type: 'bridge', subnet: '172.17.0.0/16'},
                {id: 'host', name: 'host', type: 'host', subnet: ''}
            ],
            message: `Network service not available: ${error.message}`
        });
    }
});

// Create network
router.post('/networks', async (req, res) => {
    try {
        const {name, subnet, type} = req.body;
        const node = req.query.node;
        
        if (!name) {
            return res.status(400).json({error: 'Network name is required'});
        }
        
        const args = ['network', 'create', name];
        if (subnet) args.push('--subnet', subnet);
        if (type) args.push('--type', type);
        
        const output = await execRnx(args, {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to create network:', error);
        res.status(500).json({error: error.message});
    }
});

// Delete network
router.delete('/networks/:networkName', async (req, res) => {
    try {
        const {networkName} = req.params;
        const node = req.query.node;
        
        const output = await execRnx(['network', 'delete', networkName], {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to delete network:', error);
        res.status(500).json({error: error.message});
    }
});

// Get runtimes
router.get('/runtimes', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['runtime', 'list', '--json'], {node});
        
        let runtimes = [];
        if (output && output.trim()) {
            try {
                const runtimeData = JSON.parse(output);
                // rnx runtime list returns an array directly, not an object with runtimes property
                const rawRuntimes = Array.isArray(runtimeData) ? runtimeData : (runtimeData.runtimes || []);
                
                // Transform to match UI expectations
                runtimes = rawRuntimes.map(runtime => ({
                    id: runtime.id,
                    name: runtime.name || runtime.id,
                    version: runtime.version,
                    size: runtime.size || `${Math.round(runtime.size_bytes / 1024 / 1024)}MB`,
                    description: runtime.description || `${runtime.language} runtime`
                }));
            } catch (e) {
                console.warn('Failed to parse runtime JSON:', e.message);
                runtimes = [];
            }
        }
        
        res.json({runtimes});
    } catch (error) {
        console.error('Failed to list runtimes:', error);
        res.json({runtimes: []});
    }
});

// Get monitor data (legacy endpoint)
router.get('/monitor', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['monitor', 'status', '--json'], {node});
        const metrics = JSON.parse(output);
        res.json(metrics);
    } catch (error) {
        console.error('Failed to get monitor data:', error);
        res.status(500).json({error: error.message});
    }
});

export default router;