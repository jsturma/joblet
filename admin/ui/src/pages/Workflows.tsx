import React, {useState} from 'react';
import {useWorkflows} from '../hooks/useWorkflows';
import WorkflowList from '../components/Workflow/WorkflowList';
import WorkflowDetail from '../components/Workflow/WorkflowDetail';
import {Plus, RotateCcw} from 'lucide-react';

const Workflows: React.FC = () => {
    const {
        paginatedWorkflows,
        loading,
        error,
        refreshWorkflows,
        currentPage,
        pageSize,
        totalWorkflows,
        totalPages,
        setCurrentPage,
        setPageSize,
        workflows: allWorkflows
    } = useWorkflows();
    const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(null);

    const selectedWorkflow = selectedWorkflowId
        ? allWorkflows.find(w => w.id.toString() === selectedWorkflowId)
        : null;

    const handleWorkflowClick = (workflowId: string) => {
        setSelectedWorkflowId(workflowId);
    };

    const handleBack = () => {
        setSelectedWorkflowId(null);
    };

    // Show workflow detail view if a workflow is selected
    if (selectedWorkflow) {
        return (
            <WorkflowDetail
                workflowId={selectedWorkflow.id.toString()}
                onBack={handleBack}
                onRefresh={refreshWorkflows}
            />
        );
    }

    // Show workflow list view
    return (
        <div className="p-6">
            <div className="mb-8">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-white">Workflows</h1>
                        <p className="mt-2 text-gray-300">Visual workflow management and orchestration</p>
                    </div>
                    <div className="flex space-x-3">
                        <button
                            onClick={refreshWorkflows}
                            disabled={loading}
                            className={`inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium ${
                                loading
                                    ? 'text-gray-400 bg-gray-100 cursor-not-allowed'
                                    : 'text-gray-700 bg-white hover:bg-gray-50'
                            }`}
                        >
                            <RotateCcw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`}/>
                            {loading ? 'Refreshing...' : 'Refresh'}
                        </button>
                        <button
                            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                            disabled
                        >
                            <Plus className="h-4 w-4 mr-2"/>
                            New Workflow
                        </button>
                    </div>
                </div>
            </div>

            {loading ? (
                <div className="bg-gray-800 rounded-lg shadow">
                    <div className="p-6">
                        <p className="text-white">Loading workflows...</p>
                    </div>
                </div>
            ) : error ? (
                <div className="bg-gray-800 rounded-lg shadow">
                    <div className="p-6">
                        <p className="text-red-500">Error: {error}</p>
                    </div>
                </div>
            ) : (
                <WorkflowList
                    workflows={paginatedWorkflows}
                    onWorkflowClick={handleWorkflowClick}
                    loading={loading}
                    currentPage={currentPage}
                    pageSize={pageSize}
                    totalWorkflows={totalWorkflows}
                    totalPages={totalPages}
                    setCurrentPage={setCurrentPage}
                    setPageSize={setPageSize}
                />
            )}
        </div>
    );
};

export default Workflows;