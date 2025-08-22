import {useEffect} from 'react';
import {useNode} from '../contexts/NodeContext';
import {apiService} from '../services/apiService';

interface ApiProviderProps {
    children: React.ReactNode;
}

export const ApiProvider: React.FC<ApiProviderProps> = ({children}) => {
    const {selectedNode} = useNode();

    useEffect(() => {
        apiService.setNode(selectedNode);
    }, [selectedNode]);

    return <>{children}</>;
};