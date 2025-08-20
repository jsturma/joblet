import {useCallback, useEffect, useState} from 'react';
import {apiService} from '../services/apiService';

type WorkflowStatus = 'RUNNING' | 'COMPLETED' | 'FAILED' | 'QUEUED' | 'STOPPED';

export interface Workflow {
    id: number;
    name: string;
    workflow: string;
    status: WorkflowStatus;
    total_jobs: number;
    completed_jobs: number;
    failed_jobs: number;
    created_at: string;
    started_at?: string;
    completed_at?: string;
}

interface UseWorkflowsReturn {
    workflows: Workflow[];
    loading: boolean;
    error: string | null;
    refreshWorkflows: () => Promise<void>;
    // Pagination
    currentPage: number;
    pageSize: number;
    totalWorkflows: number;
    totalPages: number;
    paginatedWorkflows: Workflow[];
    setCurrentPage: (page: number) => void;
    setPageSize: (size: number) => void;
}

export const useWorkflows = (): UseWorkflowsReturn => {
    const [workflows, setWorkflows] = useState<Workflow[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    // Pagination state
    const [currentPage, setCurrentPage] = useState<number>(1);
    const [pageSize, setPageSize] = useState<number>(10);

    const fetchWorkflows = useCallback(async (showLoading: boolean = false): Promise<void> => {
        try {
            if (showLoading) {
                setLoading(true);
            }
            setError(null);
            const response = await apiService.getWorkflows();
            // Sort workflows in descending order by ID (newest first)
            const sortedWorkflows = response.sort((a, b) => b.id - a.id);
            setWorkflows(sortedWorkflows);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch workflows');
            console.error('Failed to fetch workflows:', err);
        } finally {
            if (showLoading) {
                setLoading(false);
            }
        }
    }, []);

    const refreshWorkflows = useCallback(async (): Promise<void> => {
        await fetchWorkflows(true);
    }, [fetchWorkflows]);

    // Calculate pagination values
    const totalWorkflows = workflows.length;
    const totalPages = Math.ceil(totalWorkflows / pageSize);
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    const paginatedWorkflows = workflows.slice(startIndex, endIndex);

    // Reset to page 1 if current page is beyond available pages
    useEffect(() => {
        if (currentPage > totalPages && totalPages > 0) {
            setCurrentPage(1);
        }
    }, [currentPage, totalPages]);

    // Handle page size changes
    const handleSetPageSize = useCallback((size: number) => {
        setPageSize(size);
        setCurrentPage(1); // Reset to first page when changing page size
    }, []);

    useEffect(() => {
        // Initial load with loading indicator
        fetchWorkflows(true);

        // Poll for updates every 10 seconds (without loading indicator)
        const interval = setInterval(() => fetchWorkflows(false), 10000);
        return () => clearInterval(interval);
    }, [fetchWorkflows]);

    return {
        workflows,
        loading,
        error,
        refreshWorkflows,
        currentPage,
        pageSize,
        totalWorkflows,
        totalPages,
        paginatedWorkflows,
        setCurrentPage,
        setPageSize: handleSetPageSize
    };
};