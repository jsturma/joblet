import {createContext, ReactNode, useContext, useState} from 'react';

interface NodeContextType {
    selectedNode: string;
    setSelectedNode: (node: string) => void;
}

const NodeContext = createContext<NodeContextType | undefined>(undefined);

export const useNode = () => {
    const context = useContext(NodeContext);
    if (!context) {
        throw new Error('useNode must be used within a NodeProvider');
    }
    return context;
};

interface NodeProviderProps {
    children: ReactNode;
}

export const NodeProvider: React.FC<NodeProviderProps> = ({children}) => {
    // Load initial node from localStorage or use default
    const [selectedNode, setSelectedNodeState] = useState<string>(() => {
        const saved = localStorage.getItem('selectedNode');
        return saved || 'default';
    });

    // Save to localStorage whenever node changes
    const setSelectedNode = (node: string) => {
        setSelectedNodeState(node);
        localStorage.setItem('selectedNode', node);
    };

    return (
        <NodeContext.Provider value={{selectedNode, setSelectedNode}}>
            {children}
        </NodeContext.Provider>
    );
};