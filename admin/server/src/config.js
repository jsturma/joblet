// Configuration constants and environment setup
export const config = {
    PORT: process.env.PORT || 5173,
    BIND_ADDRESS: process.env.BIND_ADDRESS || 'localhost',
    RNX_PATH: process.env.RNX_PATH || '../../bin/rnx'
};