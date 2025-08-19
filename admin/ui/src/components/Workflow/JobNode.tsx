import React from 'react';
import {JobStatus, WorkflowJob} from '../../types/job';
import {Clock, Cpu, HardDrive} from 'lucide-react';
import clsx from 'clsx';

interface JobNodeProps {
    job: WorkflowJob;
    position: { x: number; y: number };
    selected: boolean;
    onClick: () => void;
    onDoubleClick?: () => void;
    onMouseDown?: (e: React.MouseEvent) => void;
    isDragging?: boolean;
}

const getStatusColor = (status: JobStatus): string => {
    const colors: Record<string, string> = {
        'INITIALIZING': 'border-gray-400 bg-gray-50',
        'RUNNING': 'border-yellow-400 bg-yellow-50',
        'COMPLETED': 'border-green-400 bg-green-50',
        'FAILED': 'border-red-400 bg-red-50',
        'STOPPED': 'border-gray-400 bg-gray-100',
        'QUEUED': 'border-blue-400 bg-blue-50',
        'WAITING': 'border-purple-400 bg-purple-50',
        'CANCELLED': 'border-orange-400 bg-orange-50',
        'PENDING': 'border-indigo-400 bg-indigo-50'
    };
    return colors[status] || 'border-gray-400 bg-gray-50';
};

const getStatusIcon = (status: JobStatus): string => {
    const icons: Record<string, string> = {
        'INITIALIZING': '🔄',
        'RUNNING': '🟡',
        'COMPLETED': '✅',
        'FAILED': '❌',
        'STOPPED': '⏹',
        'QUEUED': '⚪',
        'WAITING': '⏸',
        'CANCELLED': '🚫',
        'PENDING': '⏳'
    };
    return icons[status] || '⚪';
};

const formatDuration = (duration: number): string => {
    if (!duration) return '-';
    const seconds = Math.floor(duration / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
        return `${hours}h ${minutes % 60}m`;
    } else if (minutes > 0) {
        return `${minutes}m ${seconds % 60}s`;
    } else {
        return `${seconds}s`;
    }
};

export const JobNode: React.FC<JobNodeProps> = ({
                                                    job,
                                                    position,
                                                    selected,
                                                    onClick,
                                                    onDoubleClick,
                                                    onMouseDown,
                                                    isDragging
                                                }) => {
    const statusColor = getStatusColor(job.status);
    const statusIcon = getStatusIcon(job.status);

    return (
        <div
            className={clsx(
                'absolute transition-all duration-200 transform select-none',
                isDragging ? 'cursor-grabbing z-20' : 'cursor-grab',
                selected ? 'scale-105 z-10' : 'z-0 hover:scale-102'
            )}
            style={{
                left: position.x,
                top: position.y,
                transform: selected ? 'scale(1.05)' : 'scale(1)',
                userSelect: 'none',
                WebkitUserSelect: 'none',
                MozUserSelect: 'none',
                msUserSelect: 'none'
            }}
            onClick={onClick}
            onDoubleClick={onDoubleClick}
            onMouseDown={onMouseDown}
        >
            <div className={clsx(
                'bg-white rounded-lg shadow-md border-2 p-3 min-w-[140px] max-w-[200px] select-none',
                statusColor,
                selected ? 'ring-2 ring-blue-500 ring-offset-2' : ''
            )}>
                {/* Header with Job Name */}
                <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-semibold text-gray-900 truncate flex-1" title={job.name || job.id}>
            {job.name || job.id}
          </span>
                    <span className="text-lg ml-2">{statusIcon}</span>
                </div>

                {/* Status */}
                <div className="text-xs text-gray-600 mb-2 font-medium">
                    {job.status}
                </div>

                {/* Job ID (if different from name) */}
                {job.name && job.name !== job.id && (
                    <div className="text-xs text-gray-400 mb-2 truncate" title={job.id}>
                        ID: {job.id.slice(0, 8)}...
                    </div>
                )}

                {/* Duration */}
                {(job.status === 'RUNNING' || job.status === 'COMPLETED') && (
                    <div className="flex items-center text-xs text-gray-500 mb-2">
                        <Clock className="w-3 h-3 mr-1"/>
                        {formatDuration(job.duration)}
                    </div>
                )}

                {/* Resource Usage */}
                {job.resourceUsage && job.status === 'RUNNING' && (
                    <div className="space-y-1">
                        <div className="flex items-center text-xs">
                            <HardDrive className="w-3 h-3 mr-1 text-blue-500"/>
                            <div className="flex-1 bg-gray-200 rounded-full h-1.5 ml-1">
                                <div
                                    className="bg-blue-500 h-1.5 rounded-full transition-all duration-300"
                                    style={{width: `${Math.min(job.resourceUsage.memoryPercent, 100)}%`}}
                                />
                            </div>
                            <span className="ml-1 text-gray-500 text-xs">
                {Math.round(job.resourceUsage.memoryPercent)}%
              </span>
                        </div>

                        <div className="flex items-center text-xs">
                            <Cpu className="w-3 h-3 mr-1 text-green-500"/>
                            <div className="flex-1 bg-gray-200 rounded-full h-1.5 ml-1">
                                <div
                                    className="bg-green-500 h-1.5 rounded-full transition-all duration-300"
                                    style={{width: `${Math.min(job.resourceUsage.cpuPercent, 100)}%`}}
                                />
                            </div>
                            <span className="ml-1 text-gray-500 text-xs">
                {Math.round(job.resourceUsage.cpuPercent)}%
              </span>
                        </div>
                    </div>
                )}

                {/* Resource Limits for Queued/Waiting Jobs */}
                {(job.status === 'QUEUED' || job.status === 'WAITING') && (
                    <div className="text-xs text-gray-500">
                        {job.maxCPU > 0 && <div>CPU: {job.maxCPU}%</div>}
                        {job.maxMemory > 0 && <div>RAM: {job.maxMemory}MB</div>}
                        {job.cpuCores && <div>Cores: {job.cpuCores}</div>}
                    </div>
                )}
            </div>
        </div>
    );
};