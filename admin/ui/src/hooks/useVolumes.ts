import {useCallback, useEffect, useState} from 'react';
import {apiService} from '../services/apiService';

interface Volume {
    id?: string;
    name: string;
    size: string;
    type: string;
    created_time?: string;
    mountPath?: string;
}

interface UseVolumesReturn {
    volumes: Volume[];
    loading: boolean;
    error: string | null;
    refetch: () => void;
}

export const useVolumes = (): UseVolumesReturn => {
    const [volumes, setVolumes] = useState<Volume[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const fetchVolumes = useCallback(async (): Promise<void> => {
        try {
            setLoading(true);
            setError(null);
            const response = await apiService.getVolumes();
            setVolumes(response.volumes || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch volumes');
        } finally {
            setLoading(false);
        }
    }, []);

    const refetch = useCallback(() => {
        fetchVolumes();
    }, [fetchVolumes]);

    // Volumes are static resources, no auto-refresh needed

    useEffect(() => {
        fetchVolumes();
    }, [fetchVolumes]);

    return {volumes, loading, error, refetch};
};