import express from 'express';
import {execRnx} from '../utils/rnxExecutor.js';

const router = express.Router();

// List all jobs
router.get('/', async (req, res) => {
    try {
        const node = req.query.node;
        const output = await execRnx(['list', '--json'], {node});

        // Parse rnx list output
        let jobs = [];
        if (output && output.trim()) {
            try {
                jobs = JSON.parse(output);
                if (!Array.isArray(jobs)) {
                    jobs = [];
                }
            } catch (e) {
                console.warn('Failed to parse JSON from rnx list:', e.message);
                jobs = [];
            }
        }

        res.json(jobs);
    } catch (error) {
        console.error('Failed to list jobs:', error);
        res.json([]);
    }
});

// Execute a new job
router.post('/execute', async (req, res) => {
    try {
        const {command, runtime, volumes, uploads, network, cpuLimit, memoryLimit, cpuCores} = req.body;
        const node = req.query.node;

        if (!command) {
            return res.status(400).json({error: 'Command is required'});
        }

        const args = ['run'];
        
        // Add resource specifications
        if (runtime) args.push('--runtime', runtime);
        if (cpuLimit) args.push('--cpu-limit', cpuLimit);
        if (memoryLimit) args.push('--memory-limit', memoryLimit);
        if (cpuCores) args.push('--cpu-cores', cpuCores.toString());
        if (network) args.push('--network', network);

        // Add volumes
        if (volumes && volumes.length > 0) {
            volumes.forEach(volume => {
                args.push('--volume', volume);
            });
        }

        // Add uploads
        if (uploads && uploads.length > 0) {
            uploads.forEach(upload => {
                args.push('--upload', upload);
            });
        }

        // Add the command
        args.push(command);

        const output = await execRnx(args, {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to execute job:', error);
        res.status(500).json({error: error.message});
    }
});

// Get job details
router.get('/:jobId', async (req, res) => {
    try {
        const {jobId} = req.params;
        const node = req.query.node;
        const output = await execRnx(['status', jobId, '--json'], {node});

        let jobDetails;
        if (output && output.trim()) {
            try {
                const rawJob = JSON.parse(output);

                // Map RNX response to expected Job interface
                jobDetails = {
                    id: rawJob.jobUuid || rawJob.id || jobId,
                    command: rawJob.command || '',
                    args: rawJob.args || [],
                    status: rawJob.status || 'UNKNOWN',
                    startTime: rawJob.startTime || '',
                    endTime: rawJob.endTime || '',
                    duration: rawJob.duration || 0,
                    exitCode: rawJob.exitCode || rawJob.exit_code,
                    maxCPU: rawJob.maxCPU || rawJob.max_cpu || 0,
                    maxMemory: rawJob.maxMemory || rawJob.max_memory || 0,
                    maxIOBPS: rawJob.maxIOBPS || rawJob.max_iobps || 0,
                    cpuCores: rawJob.cpuCores || rawJob.cpu_cores || '',
                    runtime: rawJob.runtime || '',
                    network: rawJob.network || 'bridge',
                    volumes: rawJob.volumes || [],
                    uploads: rawJob.uploads || [],
                    uploadDirs: rawJob.uploadDirs || rawJob.upload_dirs || [],
                    envVars: rawJob.environment || {},
                    secretEnvVars: {},
                    dependsOn: rawJob.dependsOn || rawJob.depends_on || [],
                    resourceUsage: rawJob.resourceUsage || rawJob.resource_usage
                };
            } catch (e) {
                console.warn('Failed to parse JSON from rnx status:', e.message);
                jobDetails = {
                    id: jobId,
                    status: 'UNKNOWN',
                    message: output || 'No output from status command'
                };
            }
        } else {
            jobDetails = {
                id: jobId,
                status: 'NOT_FOUND',
                message: 'Job not found or no output'
            };
        }

        res.json(jobDetails);
    } catch (error) {
        console.error(`Failed to get job ${req.params.jobId}:`, error);
        res.status(500).json({
            error: 'Failed to get job details',
            message: error.message,
            id: req.params.jobId,
            status: 'ERROR'
        });
    }
});

// Stop a job
router.post('/:jobId/stop', async (req, res) => {
    try {
        const {jobId} = req.params;
        const node = req.query.node;
        const output = await execRnx(['stop', jobId], {node});
        res.json({success: true, output});
    } catch (error) {
        console.error('Failed to stop job:', error);
        res.status(500).json({error: error.message});
    }
});

export default router;