import {useCallback, useEffect, useState} from 'react';
import {Job, JobExecuteRequest} from '../types/job';
import {apiService} from '../services/apiService';

interface UseJobsReturn {
    jobs: Job[];
    loading: boolean;
    error: string | null;
    refreshJobs: () => Promise<void>;
    executeJob: (request: JobExecuteRequest) => Promise<string>;
    stopJob: (jobId: string) => Promise<void>;
    // Pagination
    currentPage: number;
    pageSize: number;
    totalJobs: number;
    totalPages: number;
    paginatedJobs: Job[];
    setCurrentPage: (page: number) => void;
    setPageSize: (size: number) => void;
}

// Helper function to extract numeric ID from job ID string
const getNumericId = (id: string): number => {
    const match = id.match(/\d+/);
    return match ? parseInt(match[0], 10) : 0;
};

// Sort jobs by startTime (oldest first), then by numeric ID in ascending order
const sortJobs = (jobs: Job[]): Job[] => {
    return [...jobs].sort((a, b) => {
        // Primary sort: by startTime (older jobs first) - check both field name variations
        const aTime = (a as any).start_time || a.startTime;
        const bTime = (b as any).start_time || b.startTime;
        if (aTime && bTime) {
            const timeA = new Date(aTime).getTime();
            const timeB = new Date(bTime).getTime();
            if (timeA !== timeB) {
                return timeA - timeB; // Ascending order (oldest first)
            }
        }

        // Secondary sort: by numeric ID for consistency
        const numA = getNumericId(a.id);
        const numB = getNumericId(b.id);

        if (numA !== numB) {
            return numA - numB; // Ascending order by numeric ID
        }

        // Fallback: string comparison if numeric parts are equal
        return a.id.localeCompare(b.id);
    });
};

export const useJobs = (): UseJobsReturn => {
    const [jobs, setJobs] = useState<Job[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    // Pagination state
    const [currentPage, setCurrentPage] = useState<number>(1);
    const [pageSize, setPageSize] = useState<number>(10);

    const fetchJobs = useCallback(async (showLoading = false): Promise<void> => {
        try {
            if (showLoading) {
                setLoading(true);
            }
            setError(null);
            const response = await apiService.getJobs();
            const sortedJobs = sortJobs(response);
            setJobs(sortedJobs);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch jobs');
            console.error('Failed to fetch jobs:', err);
        } finally {
            if (showLoading) {
                setLoading(false);
            }
        }
    }, []);

    const refreshJobs = useCallback(async (): Promise<void> => {
        await fetchJobs(true);
    }, [fetchJobs]);

    const executeJob = useCallback(async (request: JobExecuteRequest): Promise<string> => {
        try {
            const response = await apiService.executeJob(request);
            await fetchJobs(false); // Refresh job list without loading indicator
            return response.jobId;
        } catch (err) {
            throw new Error(err instanceof Error ? err.message : 'Failed to execute job');
        }
    }, [fetchJobs]);

    const stopJob = useCallback(async (jobId: string): Promise<void> => {
        try {
            await apiService.stopJob(jobId);
            await fetchJobs(false); // Refresh job list without loading indicator
        } catch (err) {
            throw new Error(err instanceof Error ? err.message : 'Failed to stop job');
        }
    }, [fetchJobs]);

    // Calculate pagination values
    const totalJobs = jobs.length;
    const totalPages = Math.ceil(totalJobs / pageSize);
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    const paginatedJobs = jobs.slice(startIndex, endIndex);

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
        fetchJobs(true);

        // Poll for updates every 5 seconds (without loading indicator)
        const interval = setInterval(() => fetchJobs(false), 5000);
        return () => clearInterval(interval);
    }, [fetchJobs]);

    return {
        jobs,
        loading,
        error,
        refreshJobs,
        executeJob,
        stopJob,
        currentPage,
        pageSize,
        totalJobs,
        totalPages,
        paginatedJobs,
        setCurrentPage,
        setPageSize: handleSetPageSize
    };
};