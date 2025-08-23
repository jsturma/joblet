// Node.js version check - ensure we have Node.js 18+
const nodeVersion = process.version;
const majorVersion = parseInt(nodeVersion.slice(1).split('.')[0]);

if (majorVersion < 18) {
    console.error(`❌ Node.js ${nodeVersion} detected, but Node.js 18+ is required`);
    console.error(`💡 Please upgrade Node.js to version 18 or later`);
    console.error(`   Visit: https://nodejs.org/`);
    process.exit(1);
}

console.log('🔄 Starting modular Joblet Admin Server...');

// Import and start the refactored modular server
import('./src/server.js').then(() => {
    console.log('✅ Modular server architecture loaded successfully');
}).catch((error) => {
    console.error('❌ Failed to start modular server:', error);
    console.error('💡 Try: npm run start:legacy for the monolithic version');
    process.exit(1);
});