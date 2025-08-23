import {spawn} from 'child_process';
import {execRnx} from '../utils/rnxExecutor.js';
import {config} from '../config.js';

export function handleLogStream(ws, jobId) {
    ws.send(JSON.stringify({
        type: 'connection',
        message: `Connected to log stream for job ${jobId}`,
        jobId: jobId,
        time: new Date().toISOString()
    }));

    // Start streaming logs using rnx log --follow
    const logProcess = spawn(config.RNX_PATH, ['log', jobId, '--follow'], {
        stdio: ['pipe', 'pipe', 'pipe']
    });

    let isAlive = true;

    // Stream stdout (logs)
    logProcess.stdout.on('data', (data) => {
        if (!isAlive) return;

        const logLines = data.toString().split('\n').filter(line => line.trim());
        logLines.forEach(line => {
            ws.send(JSON.stringify({
                type: 'log',
                jobId: jobId,
                message: line,
                time: new Date().toISOString()
            }));
        });
    });

    // Stream stderr (errors)
    logProcess.stderr.on('data', (data) => {
        if (!isAlive) return;

        const errorLines = data.toString().split('\n').filter(line => line.trim());
        errorLines.forEach(line => {
            ws.send(JSON.stringify({
                type: 'error',
                jobId: jobId,
                message: line,
                time: new Date().toISOString()
            }));
        });
    });

    // Handle process exit
    logProcess.on('close', (code) => {
        if (!isAlive) return;

        ws.send(JSON.stringify({
            type: 'connection',
            message: `Log stream for job ${jobId} ended with code ${code}`,
            jobId: jobId,
            time: new Date().toISOString()
        }));
    });

    // Cleanup on WebSocket close
    ws.on('close', () => {
        isAlive = false;
        if (logProcess && !logProcess.killed) {
            logProcess.kill();
        }
    });
}

export function handleWorkflowStatusStream(ws, workflowId, node) {
    const interval = setInterval(async () => {
        try {
            const output = await execRnx(['list', '--json'], {node});
            const jobs = JSON.parse(output);
            
            // Find jobs that belong to this workflow
            const workflowJobs = jobs.filter(job => job.workflowId === workflowId);
            
            ws.send(JSON.stringify({
                type: 'workflow_status_change',
                workflowId: workflowId,
                jobs: workflowJobs,
                time: new Date().toISOString()
            }));
        } catch (error) {
            ws.send(JSON.stringify({
                type: 'error',
                message: `Workflow status update failed: ${error.message}`,
                time: new Date().toISOString()
            }));
        }
    }, 3000);

    ws.on('close', () => {
        clearInterval(interval);
    });
}

export function handleMonitorStream(ws, node) {
    const interval = setInterval(async () => {
        try {
            const output = await execRnx(['monitor', 'status', '--json'], {node});
            const metrics = JSON.parse(output);

            ws.send(JSON.stringify({
                type: 'metrics',
                data: metrics,
                time: new Date().toISOString()
            }));
        } catch (error) {
            ws.send(JSON.stringify({
                type: 'error',
                message: `Monitor command failed: ${error.message}`,
                time: new Date().toISOString()
            }));
        }
    }, 5000);

    ws.on('close', () => {
        clearInterval(interval);
    });
}