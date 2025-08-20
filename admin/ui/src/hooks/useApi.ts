import {useNode} from '../contexts/NodeContext';

export const useApi = () => {
    const {selectedNode} = useNode();

    const apiCall = async (url: string, options: RequestInit = {}) => {
        // Add node as query parameter for GET requests
        if (!options.method || options.method === 'GET') {
            const separator = url.includes('?') ? '&' : '?';
            url = `${url}${separator}node=${encodeURIComponent(selectedNode)}`;
        }

        // Add node to body for POST/PUT requests
        if (options.method && ['POST', 'PUT', 'PATCH'].includes(options.method)) {
            if (options.body) {
                const body = JSON.parse(options.body as string);
                body.node = selectedNode;
                options.body = JSON.stringify(body);
            } else {
                options.body = JSON.stringify({node: selectedNode});
            }
        }

        return fetch(url, options);
    };

    const getWebSocketUrl = (path: string) => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const separator = path.includes('?') ? '&' : '?';
        return `${protocol}//${host}${path}${separator}node=${encodeURIComponent(selectedNode)}`;
    };

    return {apiCall, getWebSocketUrl, selectedNode};
};