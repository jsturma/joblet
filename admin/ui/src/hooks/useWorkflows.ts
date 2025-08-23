import {useCallback, useEffect, useState} from 'react';
import {apiService} from '../services/apiService';
import {useAutoRefresh} from './useAutoRefresh';

type WorkflowStatus = 'RUNNING' | 'COMPLETED' | 'FAILED' | 'QUEUED' | 'STOPPED';

export interface Workflow {
    uuid: string; // UUID field for consistency with WorkflowList
    id?: string | number; // Optional ID field for backward compatibility
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

    const fetchWorkflows = useCallback(async (showLoading = false): Promise<void> => {
        try {
            if (showLoading) {
                setLoading(true);
            }
            setError(null);
            const response = await apiService.getWorkflows();

            // Transform API response: ensure uuid field and add name field
            const transformedWorkflows = response.map(workflow => ({
                ...workflow,
                uuid: workflow.uuid || workflow.id, // Ensure uuid field exists
                name: workflow.workflow || `Workflow ${workflow.uuid ? workflow.uuid.substring(0, 8) : 'Unknown'}` // Use workflow filename as name
            }));

            // Sort workflows by creation date (newest first), fallback to UUID comparison
            const sortedWorkflows = transformedWorkflows.sort((a, b) => {
                // First try to sort by created_at timestamp
                if (a.created_at && b.created_at) {
                    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
                }
                // For UUIDs, use string comparison
                return String(b.uuid).localeCompare(String(a.uuid));
            });
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

    // Auto-refresh functionality using user settings
    useAutoRefresh(() => fetchWorkflows(false));

    useEffect(() => {
        // Initial load with loading indicator
        fetchWorkflows(true);
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