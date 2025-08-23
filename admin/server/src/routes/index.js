import nodesRouter from './nodes.js';
import jobsRouter from './jobs.js';
import workflowsRouter from './workflows.js';
import systemRouter from './system.js';

export function setupRoutes(app) {
    app.use('/api/nodes', nodesRouter);
    app.use('/api/jobs', jobsRouter);
    app.use('/api/workflows', workflowsRouter);
    app.use('/api', systemRouter);
}