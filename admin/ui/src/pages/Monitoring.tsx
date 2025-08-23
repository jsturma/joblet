// React import not needed with modern JSX transform
import {useMonitorStream} from '../hooks/useMonitorStream';
import {useSystemInfo} from '../hooks/useSystemInfo';
import HostInfoCard from '../components/Monitoring/HostInfoCard';
import CPUDetailsCard from '../components/Monitoring/CPUDetailsCard';
import MemoryDetailsCard from '../components/Monitoring/MemoryDetailsCard';
import DisksCard from '../components/Monitoring/DisksCard';
import NetworkCard from '../components/Monitoring/NetworkCard';
import ProcessesCard from '../components/Monitoring/ProcessesCard';

const Monitoring: React.FC = () => {
    const {metrics, connected, error: metricsError} = useMonitorStream();
    const {systemInfo, loading: systemLoading, error: systemError} = useSystemInfo();

    const loading = systemLoading;
    const error = metricsError || systemError;

    return (
        <div className="p-6">
            <div className="mb-8">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-white">System Monitoring</h1>
                        <p className="mt-2 text-gray-300">Real-time system metrics and performance</p>
                        <div className="mt-2 flex items-center text-sm">
                            <div className={`w-2 h-2 rounded-full mr-2 ${connected ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`}></div>
                            <span className="text-gray-400">
                                {connected ? 'Live Updates' : 'Connecting...'}
                            </span>
                        </div>
                    </div>
                </div>
            </div>

            {loading ? (
                <div className="bg-gray-800 rounded-lg shadow p-6">
                    <p className="text-gray-300">Loading system information...</p>
                </div>
            ) : error ? (
                <div className="bg-gray-800 rounded-lg shadow p-6">
                    <p className="text-red-400">Error: {error}</p>
                </div>
            ) : (
                <div className="space-y-6">
                    {/* Host Information */}
                    {systemInfo?.hostInfo && (
                        <HostInfoCard hostInfo={systemInfo.hostInfo}/>
                    )}

                    {/* CPU Details */}
                    {systemInfo?.cpuInfo && (
                        <CPUDetailsCard 
                            cpuInfo={{
                                ...systemInfo.cpuInfo,
                                ...(metrics?.cpu && {
                                    usage: metrics.cpu.usagePercent,
                                    loadAverage: metrics.cpu.loadAverage,
                                    perCoreUsage: metrics.cpu.perCoreUsage
                                })
                            }}
                        />
                    )}

                    {/* Memory Details */}
                    {systemInfo?.memoryInfo && (
                        <MemoryDetailsCard 
                            memoryInfo={{
                                ...systemInfo.memoryInfo,
                                ...(metrics?.memory && {
                                    used: metrics.memory.usedBytes,
                                    available: metrics.memory.availableBytes,
                                    total: metrics.memory.totalBytes,
                                    percent: metrics.memory.usagePercent,
                                    buffers: metrics.memory.bufferedBytes,
                                    cached: metrics.memory.cachedBytes
                                })
                            }}
                        />
                    )}

                    {/* Disk Information */}
                    {systemInfo?.disksInfo && (
                        <DisksCard disksInfo={systemInfo.disksInfo}/>
                    )}

                    {/* Network Interfaces */}
                    {systemInfo?.networkInfo && (
                        <NetworkCard networkInfo={systemInfo.networkInfo}/>
                    )}

                    {/* Running Processes */}
                    {systemInfo?.processesInfo && (
                        <ProcessesCard processesInfo={systemInfo.processesInfo}/>
                    )}

                </div>
            )}
        </div>
    );
};

export default Monitoring;