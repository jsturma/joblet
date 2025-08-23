// React import not needed with modern JSX transform
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
        'INITIALIZING': 'border-gray-500 bg-gray-800',
        'RUNNING': 'border-yellow-400 bg-yellow-900',
        'COMPLETED': 'border-green-400 bg-green-900',
        'FAILED': 'border-red-400 bg-red-900',
        'STOPPED': 'border-gray-500 bg-gray-700',
        'QUEUED': 'border-blue-400 bg-blue-900',
        'WAITING': 'border-purple-400 bg-purple-900',
        'CANCELLED': 'border-orange-400 bg-orange-900',
        'PENDING': 'border-indigo-400 bg-indigo-900'
    };
    return colors[status] || 'border-gray-500 bg-gray-800';
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
                'rounded-lg shadow-lg border-2 p-2 min-w-[120px] max-w-[160px] select-none',
                statusColor,
                selected ? 'ring-2 ring-blue-400 ring-offset-1 ring-offset-gray-900' : ''
            )}>
                {/* Header with Job Name */}
                <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-semibold text-white truncate flex-1" title={job.name || job.id}>
            {job.name || job.id}
          </span>
                    <span className="text-sm ml-1">{statusIcon}</span>
                </div>

                {/* Status */}
                <div className="text-xs text-gray-300 mb-1 font-medium">
                    {job.status}
                </div>

                {/* Job ID (if different from name) */}
                {job.name && job.name !== job.id && (
                    <div className="text-xs text-gray-400 mb-1 truncate" title={job.id}>
                        {job.id.slice(0, 8)}...
                    </div>
                )}

                {/* Duration */}
                {(job.status === 'RUNNING' || job.status === 'COMPLETED') && (
                    <div className="flex items-center text-xs text-gray-400 mb-1">
                        <Clock className="w-3 h-3 mr-1"/>
                        {formatDuration(job.duration)}
                    </div>
                )}

                {/* Resource Usage */}
                {job.resourceUsage && job.status === 'RUNNING' && (
                    <div className="space-y-1">
                        <div className="flex items-center text-xs">
                            <HardDrive className="w-3 h-3 mr-1 text-blue-400"/>
                            <div className="flex-1 bg-gray-600 rounded-full h-1.5 ml-1">
                                <div
                                    className="bg-blue-400 h-1.5 rounded-full transition-all duration-300"
                                    style={{width: `${Math.min(job.resourceUsage.memoryPercent, 100)}%`}}
                                />
                            </div>
                            <span className="ml-1 text-gray-300 text-xs">
                {Math.round(job.resourceUsage.memoryPercent)}%
              </span>
                        </div>

                        <div className="flex items-center text-xs">
                            <Cpu className="w-3 h-3 mr-1 text-green-400"/>
                            <div className="flex-1 bg-gray-600 rounded-full h-1.5 ml-1">
                                <div
                                    className="bg-green-400 h-1.5 rounded-full transition-all duration-300"
                                    style={{width: `${Math.min(job.resourceUsage.cpuPercent, 100)}%`}}
                                />
                            </div>
                            <span className="ml-1 text-gray-300 text-xs">
                {Math.round(job.resourceUsage.cpuPercent)}%
              </span>
                        </div>
                    </div>
                )}

                {/* Resource Limits for Queued/Waiting Jobs */}
                {(job.status === 'QUEUED' || job.status === 'WAITING') && (
                    <div className="text-xs text-gray-400">
                        {job.maxCPU > 0 && <div>CPU: {job.maxCPU}%</div>}
                        {job.maxMemory > 0 && <div>RAM: {job.maxMemory}MB</div>}
                        {job.cpuCores && <div>Cores: {job.cpuCores}</div>}
                    </div>
                )}
            </div>
        </div>
    );
};