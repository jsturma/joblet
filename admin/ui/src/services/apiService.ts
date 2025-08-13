import {Job, JobExecuteRequest, SystemMetrics, WorkflowTemplate} from '../types';

interface Volume {
    id: string;
    name: string;
    size: number;
    mountPath: string;
}

interface Network {
    id: string;
    name: string;
    type: string;
    subnet: string;
}

interface Node {
    name: string;
    status: string;
}

export const API_BASE_URL = (import.meta as any).env?.DEV
    ? 'http://localhost:5173'
    : '';

class APIService {
    private baseURL: string;
    private currentNode: string = 'default';

    constructor() {
        this.baseURL = `${API_BASE_URL}/api`;
    }
    
    setNode(node: string) {
        this.currentNode = node;
    }
    
    // Node operations
    async getNodes(): Promise<Node[]> {
        return this.request<Node[]>('/nodes', {}, false); // Don't add node param for this call
    }

    // Job operations
    async getJobs(): Promise<Job[]> {
        return this.request<Job[]>('/jobs');
    }

    async getJob(jobId: string): Promise<Job> {
        return this.request<Job>(`/jobs/${jobId}`);
    }

    async executeJob(request: JobExecuteRequest): Promise<{ jobId: string }> {
        return this.request<{ jobId: string }>('/jobs/execute', {
            method: 'POST',
            body: JSON.stringify({ ...request, node: this.currentNode }),
        });
    }

    async stopJob(jobId: string): Promise<void> {
        await this.request(`/jobs/${jobId}/stop`, {
            method: 'POST',
            body: JSON.stringify({ node: this.currentNode }),
        });
    }

    async getJobLogs(jobId: string): Promise<{ logs: string[] }> {
        return this.request<{ logs: string[] }>(`/jobs/${jobId}/logs`);
    }

    // System monitoring
    async getSystemMetrics(): Promise<SystemMetrics> {
        return this.request<SystemMetrics>('/monitor');
    }

    // Volume operations
    async getVolumes(): Promise<{ volumes: Volume[] }> {
        return this.request<{ volumes: Volume[] }>('/volumes');
    }

    // Network operations
    async getNetworks(): Promise<{ networks: Network[] }> {
        return this.request<{ networks: Network[] }>('/networks');
    }

    // Template operations
    async executeTemplate(template: WorkflowTemplate): Promise<{ workflowId: string }> {
        return this.request<{ workflowId: string }>('/template/execute', {
            method: 'POST',
            body: JSON.stringify(template),
        });
    }

    async validateTemplate(template: WorkflowTemplate): Promise<{ valid: boolean; errors?: string[] }> {
        return this.request<{ valid: boolean; errors?: string[] }>('/template/validate', {
            method: 'POST',
            body: JSON.stringify(template),
        });
    }

    private async request<T>(
        endpoint: string,
        options: RequestInit = {},
        includeNode: boolean = true
    ): Promise<T> {
        // Add node parameter to GET requests
        let url = `${this.baseURL}${endpoint}`;
        if (includeNode && (!options.method || options.method === 'GET')) {
            const separator = url.includes('?') ? '&' : '?';
            url = `${url}${separator}node=${encodeURIComponent(this.currentNode)}`;
        }
        
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`API Error: ${response.status} - ${errorText}`);
        }

        return response.json();
    }
}

export const apiService = new APIService();