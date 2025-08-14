import React, { useEffect, useState } from 'react';
import { Cpu, HardDrive, Network, RefreshCw, Plus, Settings, Info, Trash2, X } from 'lucide-react';
import { apiService } from '../services/apiService';

interface Volume {
    id?: string;
    name: string;
    size: string;
    type: string;
    created_time?: string;
    mountPath?: string;
}

interface NetworkResource {
    id: string;
    name: string;
    type: string;
    subnet: string;
}

interface Runtime {
    id: string;
    name: string;
    version: string;
    size: string;
    description: string;
}

const Resources: React.FC = () => {
    const [volumes, setVolumes] = useState<Volume[]>([]);
    const [networks, setNetworks] = useState<NetworkResource[]>([]);
    const [runtimes, setRuntimes] = useState<Runtime[]>([]);
    const [loading, setLoading] = useState({
        volumes: true,
        networks: true,
        runtimes: true
    });
    const [error, setError] = useState({
        volumes: '',
        networks: '',
        runtimes: ''
    });
    const [deleteConfirm, setDeleteConfirm] = useState<{
        show: boolean;
        volumeName: string;
        deleting: boolean;
    }>({
        show: false,
        volumeName: '',
        deleting: false
    });

    const fetchVolumes = async () => {
        try {
            setLoading(prev => ({ ...prev, volumes: true }));
            setError(prev => ({ ...prev, volumes: '' }));
            const response = await apiService.getVolumes();
            setVolumes(response.volumes || []);
        } catch (err) {
            setError(prev => ({ ...prev, volumes: err instanceof Error ? err.message : 'Failed to fetch volumes' }));
        } finally {
            setLoading(prev => ({ ...prev, volumes: false }));
        }
    };

    const fetchNetworks = async () => {
        try {
            setLoading(prev => ({ ...prev, networks: true }));
            setError(prev => ({ ...prev, networks: '' }));
            const response = await apiService.getNetworks();
            setNetworks(response.networks || []);
        } catch (err) {
            setError(prev => ({ ...prev, networks: err instanceof Error ? err.message : 'Failed to fetch networks' }));
        } finally {
            setLoading(prev => ({ ...prev, networks: false }));
        }
    };

    const fetchRuntimes = async () => {
        try {
            setLoading(prev => ({ ...prev, runtimes: true }));
            setError(prev => ({ ...prev, runtimes: '' }));
            const response = await apiService.getRuntimes();
            setRuntimes(response.runtimes || []);
        } catch (err) {
            setError(prev => ({ ...prev, runtimes: err instanceof Error ? err.message : 'Failed to fetch runtimes' }));
        } finally {
            setLoading(prev => ({ ...prev, runtimes: false }));
        }
    };

    const refreshAll = () => {
        fetchVolumes();
        fetchNetworks();
        fetchRuntimes();
    };

    const handleDeleteVolume = async (volumeName: string) => {
        setDeleteConfirm({ show: true, volumeName, deleting: false });
    };

    const confirmDeleteVolume = async () => {
        if (!deleteConfirm.volumeName) return;
        
        setDeleteConfirm(prev => ({ ...prev, deleting: true }));
        
        try {
            await apiService.deleteVolume(deleteConfirm.volumeName);
            setDeleteConfirm({ show: false, volumeName: '', deleting: false });
            await fetchVolumes(); // Refresh the volume list
        } catch (error) {
            console.error('Failed to delete volume:', error);
            setError(prev => ({ 
                ...prev, 
                volumes: error instanceof Error ? error.message : 'Failed to delete volume' 
            }));
            setDeleteConfirm(prev => ({ ...prev, deleting: false }));
        }
    };

    const cancelDeleteVolume = () => {
        setDeleteConfirm({ show: false, volumeName: '', deleting: false });
    };

    useEffect(() => {
        refreshAll();
    }, []);

    const formatSize = (size: string | number): string => {
        if (typeof size === 'string') return size; // Already formatted
        if (size === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(size) / Math.log(k));
        return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <div className="p-6">
            <div className="mb-8">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-white">Resources</h1>
                        <p className="mt-2 text-gray-300">Manage volumes, networks, and runtime environments</p>
                    </div>
                    <button
                        onClick={refreshAll}
                        className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                    >
                        <RefreshCw className="h-4 w-4 mr-2"/>
                        Refresh All
                    </button>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Volumes */}
                <div className="bg-gray-800 rounded-lg shadow">
                    <div className="p-6">
                        <div className="flex items-center justify-between mb-4">
                            <div className="flex items-center">
                                <HardDrive className="h-6 w-6 text-blue-600 mr-3"/>
                                <h3 className="text-lg font-semibold text-gray-200">Volumes</h3>
                            </div>
                            <button
                                onClick={fetchVolumes}
                                className="text-gray-400 hover:text-gray-600"
                                title="Refresh volumes"
                            >
                                <RefreshCw className="h-4 w-4"/>
                            </button>
                        </div>

                        {loading.volumes ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500">Loading volumes...</p>
                            </div>
                        ) : error.volumes ? (
                            <div className="text-center py-8">
                                <p className="text-red-500 text-sm">{error.volumes}</p>
                            </div>
                        ) : volumes.length === 0 ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500 mb-4">No volumes configured</p>
                                <button
                                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">
                                    <Plus className="h-4 w-4 mr-2"/>
                                    Create Volume
                                </button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {volumes.map((volume, index) => (
                                    <div key={volume.id || volume.name || index} className="border rounded-lg p-3">
                                        <div className="flex items-center justify-between">
                                            <div className="flex-1">
                                                <p className="font-medium text-gray-300">{volume.name}</p>
                                                <p className="text-sm text-gray-500">{volume.type}</p>
                                                <p className="text-sm text-gray-500">{volume.mountPath || `/volumes/${volume.name}`}</p>
                                            </div>
                                            <div className="text-right mr-3">
                                                <p className="text-sm text-gray-600">{formatSize(volume.size)}</p>
                                                {volume.created_time && (
                                                    <p className="text-xs text-gray-400">{new Date(volume.created_time).toLocaleDateString()}</p>
                                                )}
                                            </div>
                                            <div>
                                                <button
                                                    onClick={() => handleDeleteVolume(volume.name)}
                                                    className="text-red-400 hover:text-red-300 p-1 rounded transition-colors"
                                                    title="Delete volume"
                                                >
                                                    <Trash2 className="h-4 w-4"/>
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                                <button
                                    className="w-full mt-4 inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                                    <Settings className="h-4 w-4 mr-2"/>
                                    Manage Volumes
                                </button>
                            </div>
                        )}
                    </div>
                </div>

                {/* Networks */}
                <div className="bg-gray-800 rounded-lg shadow">
                    <div className="p-6">
                        <div className="flex items-center justify-between mb-4">
                            <div className="flex items-center">
                                <Network className="h-6 w-6 text-green-600 mr-3"/>
                                <h3 className="text-lg font-semibold text-gray-200">Networks</h3>
                            </div>
                            <button
                                onClick={fetchNetworks}
                                className="text-gray-400 hover:text-gray-600"
                                title="Refresh networks"
                            >
                                <RefreshCw className="h-4 w-4"/>
                            </button>
                        </div>

                        {loading.networks ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500">Loading networks...</p>
                            </div>
                        ) : error.networks ? (
                            <div className="text-center py-8">
                                <p className="text-red-500 text-sm">{error.networks}</p>
                            </div>
                        ) : networks.length === 0 ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500 mb-4">No networks configured</p>
                                <button
                                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700">
                                    <Plus className="h-4 w-4 mr-2"/>
                                    Create Network
                                </button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {networks.map((network, index) => (
                                    <div key={network.id || network.name || index} className="border rounded-lg p-3">
                                        <div className="flex items-center justify-between">
                                            <div>
                                                <p className="font-medium text-gray-300">{network.name}</p>
                                                <p className="text-sm text-gray-500">{network.type}</p>
                                            </div>
                                            <div className="text-right">
                                                <p className="text-sm text-gray-600">{network.subnet || 'N/A'}</p>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                                <button
                                    className="w-full mt-4 inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                                    <Settings className="h-4 w-4 mr-2"/>
                                    Manage Networks
                                </button>
                            </div>
                        )}
                    </div>
                </div>

                {/* Runtimes */}
                <div className="bg-gray-800 rounded-lg shadow">
                    <div className="p-6">
                        <div className="flex items-center justify-between mb-4">
                            <div className="flex items-center">
                                <Cpu className="h-6 w-6 text-purple-600 mr-3"/>
                                <h3 className="text-lg font-semibold text-gray-200">Runtimes</h3>
                            </div>
                            <button
                                onClick={fetchRuntimes}
                                className="text-gray-400 hover:text-gray-600"
                                title="Refresh runtimes"
                            >
                                <RefreshCw className="h-4 w-4"/>
                            </button>
                        </div>

                        {loading.runtimes ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500">Loading runtimes...</p>
                            </div>
                        ) : error.runtimes ? (
                            <div className="text-center py-8">
                                <p className="text-red-500 text-sm">{error.runtimes}</p>
                            </div>
                        ) : runtimes.length === 0 ? (
                            <div className="text-center py-8">
                                <p className="text-gray-500 mb-4">No runtimes available</p>
                                <button
                                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-purple-600 hover:bg-purple-700">
                                    <Info className="h-4 w-4 mr-2"/>
                                    View Documentation
                                </button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {runtimes.map((runtime, index) => (
                                    <div key={runtime.id || runtime.name || index} className="border rounded-lg p-3">
                                        <div>
                                            <p className="font-medium text-gray-300">{runtime.name}</p>
                                            <p className="text-sm text-gray-500">{runtime.description}</p>
                                            <div className="flex items-center justify-between mt-2">
                                                <p className="text-xs text-gray-400">v{runtime.version}</p>
                                                <p className="text-xs text-gray-400">{runtime.size}</p>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                                <button
                                    className="w-full mt-4 inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                                    <Info className="h-4 w-4 mr-2"/>
                                    View All Runtimes
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Delete Confirmation Dialog */}
            {deleteConfirm.show && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50 flex items-center justify-center">
                    <div className="relative bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
                        <div className="p-6">
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="text-lg font-medium text-gray-200">
                                    Delete Volume
                                </h3>
                                <button
                                    onClick={cancelDeleteVolume}
                                    className="text-gray-400 hover:text-gray-300"
                                    disabled={deleteConfirm.deleting}
                                >
                                    <X className="h-5 w-5"/>
                                </button>
                            </div>
                            
                            <div className="mb-6">
                                <p className="text-gray-300 mb-2">
                                    Are you sure you want to delete the volume "{deleteConfirm.volumeName}"?
                                </p>
                                <p className="text-sm text-red-400">
                                    This action cannot be undone. All data in this volume will be permanently lost.
                                </p>
                            </div>
                            
                            <div className="flex space-x-3 justify-end">
                                <button
                                    onClick={cancelDeleteVolume}
                                    disabled={deleteConfirm.deleting}
                                    className="px-4 py-2 border border-gray-600 rounded-md text-sm font-medium text-gray-300 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Cancel
                                </button>
                                <button
                                    onClick={confirmDeleteVolume}
                                    disabled={deleteConfirm.deleting}
                                    className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-800 text-white rounded-md text-sm font-medium disabled:cursor-not-allowed flex items-center"
                                >
                                    {deleteConfirm.deleting ? (
                                        <>
                                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                                            Deleting...
                                        </>
                                    ) : (
                                        <>
                                            <Trash2 className="h-4 w-4 mr-2"/>
                                            Delete
                                        </>
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Resources;