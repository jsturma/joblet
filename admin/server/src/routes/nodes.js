import express from 'express';
import {execRnx} from '../utils/rnxExecutor.js';

const router = express.Router();

// List available nodes
router.get('/', async (req, res) => {
    try {
        const output = await execRnx(['nodes', '--json']);

        let nodes = [];
        if (output && output.trim()) {
            try {
                nodes = JSON.parse(output);
                if (!Array.isArray(nodes)) {
                    nodes = [];
                }
            } catch (e) {
                console.warn('Failed to parse JSON from rnx nodes:', e.message);
                nodes = [];
            }
        }

        // Always include default node if not present
        if (!nodes.find(n => n.name === 'default')) {
            nodes.unshift({name: 'default', status: 'active', default: true});
        }

        res.json(nodes);
    } catch (error) {
        console.error('Failed to list nodes:', error);
        // Return default node on error
        res.json([{name: 'default', status: 'active', default: true}]);
    }
});

export default router;