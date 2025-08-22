import {useEffect, useState} from 'react';
import {AlertCircle, Server} from 'lucide-react';

interface Node {
    name: string;
    status: string;
}

interface NodeSelectorProps {
    selectedNode: string;
    onNodeChange: (node: string) => void;
}

const NodeSelector: React.FC<NodeSelectorProps> = ({selectedNode, onNodeChange}) => {
    const [nodes, setNodes] = useState<Node[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetchNodes();
    }, []);

    const fetchNodes = async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await fetch('/api/nodes');
            if (!response.ok) {
                throw new Error('Failed to fetch nodes');
            }
            const data = await response.json();
            setNodes(data);
        } catch (err) {
            console.error('Error fetching nodes:', err);
            setError('Failed to load nodes');
            // Set default node on error
            setNodes([{name: 'default', status: 'active'}]);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex items-center space-x-2">
            <Server className="h-5 w-5 text-gray-500"/>
            <select
                value={selectedNode}
                onChange={(e) => onNodeChange(e.target.value)}
                className="block w-48 px-3 py-1.5 text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                disabled={loading}
            >
                {loading ? (
                    <option>Loading nodes...</option>
                ) : error ? (
                    <option value="default">Default Node</option>
                ) : (
                    nodes.map((node) => (
                        <option key={node.name} value={node.name}>
                            {node.name} {node.status === 'active' ? '✓' : '⚠'}
                        </option>
                    ))
                )}
            </select>
            {error && (
                <div title={error}>
                    <AlertCircle className="h-4 w-4 text-yellow-500"/>
                </div>
            )}
        </div>
    );
};

export default NodeSelector;