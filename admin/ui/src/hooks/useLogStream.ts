import {useCallback, useEffect, useRef, useState} from 'react';
import {useNode} from '../contexts/NodeContext';

interface LogEntry {
    message: string;
    type: 'system' | 'info' | 'output' | 'error' | 'connection';
    timestamp: string;
}

interface UseLogStreamReturn {
    logs: LogEntry[];
    connected: boolean;
    error: string | null;
    clearLogs: () => void;
}

export const useLogStream = (jobId: string | null): UseLogStreamReturn => {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [connected, setConnected] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const {selectedNode} = useNode();

    const clearLogs = useCallback(() => {
        setLogs([]);
    }, []);

    useEffect(() => {
        if (!jobId) {
            setConnected(false);
            setLogs([]);
            return;
        }

        const wsUrl = `ws://${window.location.host}/ws/logs/${jobId}?node=${encodeURIComponent(selectedNode)}`;
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
            setConnected(true);
            setError(null);
        };

        ws.onmessage = (event: MessageEvent) => {
            try {
                const logEntry = JSON.parse(event.data);
                const timestamp = new Date().toLocaleTimeString();

                let type: 'system' | 'info' | 'output' | 'error' | 'connection' = 'output';
                let message = logEntry.message;

                if (logEntry.type === 'log') {
                    if (logEntry.subtype === 'system') {
                        type = 'system';
                    } else if (logEntry.subtype === 'info') {
                        type = 'info';
                    } else {
                        type = 'output';
                    }
                } else if (logEntry.type === 'error') {
                    type = 'error';
                    message = `ERROR: ${logEntry.message}`;
                } else if (logEntry.type === 'connection') {
                    type = 'connection';
                } else if (logEntry.type === 'status') {
                    type = 'connection';
                    message = `STATUS: ${logEntry.message}`;
                } else {
                    // Fallback for unknown message types
                    message = logEntry.message || JSON.stringify(logEntry);
                }

                setLogs(prev => [...prev, {
                    message,
                    type,
                    timestamp
                }]);
            } catch {
                // Fallback for plain text logs
                const timestamp = new Date().toLocaleTimeString();
                setLogs(prev => [...prev, {
                    message: event.data,
                    type: 'output',
                    timestamp
                }]);
            }
        };

        ws.onerror = () => {
            setError('WebSocket connection error');
        };

        ws.onclose = () => {
            setConnected(false);
        };

        return () => {
            ws.close();
            wsRef.current = null;
        };
    }, [jobId, selectedNode]);

    return {logs, connected, error, clearLogs};
};