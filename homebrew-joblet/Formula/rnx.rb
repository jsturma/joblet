class Rnx < Formula
  desc "Cross-platform CLI for Joblet distributed job execution system"
  homepage "https://github.com/ehsaniara/joblet"
  license "MIT"

  version_scheme 1

  # Multi-architecture support with admin UI
  # NOTE: These URLs and checksums will be automatically updated by the auto-update workflow
  # when new releases are published. For initial testing, use actual release URLs.
  if Hardware::CPU.intel?
    url "https://github.com/ehsaniara/joblet/releases/download/v0.1.0/rnx-v0.1.0-darwin-amd64.tar.gz"
    sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  else
    url "https://github.com/ehsaniara/joblet/releases/download/v0.1.0/rnx-v0.1.0-darwin-arm64.tar.gz"
    sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  end

  # Optional Node.js dependency for admin UI
  depends_on "node" => :optional

  # Installation options
  option "with-admin", "Install with web admin UI (requires Node.js)"
  option "without-admin", "Install CLI only"

  def install
    # Install base rnx binary
    bin.install "rnx"

    # Determine installation type
    install_admin = determine_admin_installation

    if install_admin
      setup_admin_ui
    else
      ohai "✅ RNX CLI installed successfully!"
      ohai "💡 To install the web admin UI later, run:"
      ohai "   brew reinstall rnx --with-admin"
    end

    # Create bash completion
    generate_completions_from_executable(bin/"rnx", "completion", "bash")
    generate_completions_from_executable(bin/"rnx", "completion", "zsh")
    generate_completions_from_executable(bin/"rnx", "completion", "fish")
  end

  def test
    assert_match version.to_s, shell_output("#{bin}/rnx --version")
    
    # Test admin UI if installed
    admin_dir = prefix/"admin"
    if admin_dir.exist?
      assert admin_dir.join("server/package.json").exist?, "Admin server package.json should exist"
      assert admin_dir.join("ui/dist/index.html").exist?, "Admin UI build should exist"
    end
  end

  private

  def determine_admin_installation
    # Check for explicit options first
    return true if build.with? "admin"
    return false if build.with? "without-admin"

    # Auto-detection and interactive prompt
    node_available = system("which", "node", out: File::NULL, err: File::NULL)
    
    if node_available
      node_version = `node --version 2>/dev/null`.strip
      ohai "✅ Node.js detected: #{node_version}"
      
      # Verify Node.js version compatibility
      verify_nodejs_version
      
      # Interactive prompt with default Yes
      print "🤔 Would you like to install the web admin UI? (Y/n): "
      response = STDIN.gets.strip.downcase
      return response.empty? || response.start_with?("y")
    else
      ohai "❌ Node.js not detected"
      
      # Interactive prompt with default No
      print "🤔 Would you like to install Node.js and the web admin UI? (y/N): "
      response = STDIN.gets.strip.downcase
      
      if response.start_with?("y")
        # Install Node.js as a dependency
        Formula["node"].install
        # Verify Node.js version after installation
        verify_nodejs_version
        return true
      else
        return false
      end
    end
  rescue => e
    # Fallback to CLI-only installation if anything goes wrong
    opoo "Installation prompt failed: #{e.message}"
    opoo "Falling back to CLI-only installation"
    false
  end

  def setup_admin_ui
    ohai "🔧 Setting up admin UI..."
    
    # Create admin directory structure
    admin_dir = prefix/"admin"
    admin_dir.mkpath
    
    # Install admin files from release archive
    ohai "📁 Installing admin UI files..."
    cp_r "admin/.", admin_dir
    
    # Install Node.js dependencies for server
    ohai "📦 Installing admin server dependencies..."
    cd admin_dir/"server" do
      system "npm", "install", "--production", "--silent"
      unless $?.success?
        onoe "Failed to install admin server dependencies"
        raise "Admin UI setup failed"
      end
    end
    
    # Verify admin UI build exists
    ui_build = admin_dir/"ui/dist/index.html"
    unless ui_build.exist?
      onoe "Admin UI build not found"
      raise "Admin UI setup failed"
    end
    
    # Create admin launcher script
    create_admin_launcher
    
    ohai "✅ Admin UI setup complete!"
    ohai "🚀 Usage:"
    ohai "   CLI: rnx --help"
    ohai "   Web UI: rnx admin"
    
  rescue => e
    onoe "Admin UI setup failed: #{e.message}"
    opoo "Continuing with CLI-only installation..."
  end

  def create_admin_launcher
    # Create rnx-admin wrapper script
    admin_script = bin/"rnx-admin"
    admin_script.write <<~EOS
      #!/bin/bash
      # RNX Admin UI Launcher
      ADMIN_DIR="#{prefix}/admin"
      
      if [ ! -d "$ADMIN_DIR" ]; then
        echo "❌ Admin UI not installed. Install with: brew reinstall rnx --with-admin"
        exit 1
      fi
      
      echo "🚀 Starting RNX Admin UI..."
      echo "📍 Admin server directory: $ADMIN_DIR/server"
      echo "🌐 Opening http://localhost:5173 in your browser..."
      
      # Start the admin server
      cd "$ADMIN_DIR/server"
      node server.js &
      
      # Wait a moment for server to start
      sleep 2
      
      # Open browser
      open "http://localhost:5173" 2>/dev/null || echo "💡 Navigate to http://localhost:5173 in your browser"
      
      echo "✅ Admin UI is running!"
      echo "🛑 Press Ctrl+C to stop"
      
      # Keep script running
      wait
    EOS
    admin_script.chmod 0755
  end

  def verify_nodejs_version
    return unless system("which", "node", out: File::NULL, err: File::NULL)
    
    node_version = `node --version 2>/dev/null`.strip
    return if node_version.empty?
    
    # Parse version (remove 'v' prefix and get major version)
    version_match = node_version.match(/^v?(\d+)\./)
    return unless version_match
    
    major_version = version_match[1].to_i
    
    if major_version < 18
      onoe "Node.js #{node_version} detected, but admin UI requires Node.js 18+"
      ohai "💡 Please upgrade Node.js:"
      ohai "   brew upgrade node"
      ohai "   or visit: https://nodejs.org/"
      raise "Node.js version incompatible"
    end
    
    ohai "✅ Node.js #{node_version} is compatible"
  end

  def caveats
    admin_dir = prefix/"admin"
    
    if admin_dir.exist?
      <<~EOS
        🎉 RNX with Admin UI installed successfully!
        
        📋 Usage:
          CLI commands:     rnx --help
          Web Admin UI:     rnx admin  (or rnx-admin)
        
        🔧 Configuration:
          Copy your rnx-config.yml to: ~/.rnx/
        
        🌐 Admin UI will be available at: http://localhost:5173
        
        💡 Tips:
          - The admin UI provides a visual interface for job management
          - All CLI commands are still available alongside the web interface
          - Admin UI spawns local rnx CLI commands to communicate with joblet servers
      EOS
    else
      <<~EOS
        📱 RNX CLI installed successfully!
        
        📋 Usage: rnx --help
        
        🔧 Configuration:
          Copy your rnx-config.yml to: ~/.rnx/
        
        🌐 To install the web admin UI:
          brew reinstall rnx --with-admin
        
        💡 The web admin UI provides a visual interface for managing jobs,
           viewing logs, and monitoring your joblet servers.
      EOS
    end
  end
end