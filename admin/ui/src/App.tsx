import {BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import Layout from './components/Layout/Layout';
import Dashboard from './pages/Dashboard';
import Jobs from './pages/Jobs';
import Workflows from './pages/Workflows';
import Monitoring from './pages/Monitoring';
import Resources from './pages/Resources';
import {NodeProvider} from './contexts/NodeContext';
import {ApiProvider} from './providers/ApiProvider';
import {SettingsProvider} from './contexts/SettingsContext';

function App() {
    return (
        <NodeProvider>
            <SettingsProvider>
                <ApiProvider>
                    <Router>
                        <Layout>
                            <Routes>
                                <Route path="/" element={<Dashboard/>}/>
                                <Route path="/jobs" element={<Jobs/>}/>
                                <Route path="/workflows" element={<Workflows/>}/>
                                <Route path="/monitoring" element={<Monitoring/>}/>
                                <Route path="/resources" element={<Resources/>}/>
                            </Routes>
                        </Layout>
                    </Router>
                </ApiProvider>
            </SettingsProvider>
        </NodeProvider>
    );
}

export default App;