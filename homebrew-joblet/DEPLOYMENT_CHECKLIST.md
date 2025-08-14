# RNX Homebrew Deployment Checklist

Complete checklist for deploying the RNX Homebrew tap and ensuring everything works correctly.

## 🚀 Pre-Deployment Setup

### 1. Repository Creation

- [ ] **Create GitHub repository**: `ehsaniara/homebrew-joblet`
- [ ] **Set repository to public**
- [ ] **Add repository description**: "Homebrew tap for RNX - Remote job execution CLI for Joblet"
- [ ] **Set homepage**: https://github.com/ehsaniara/joblet

### 2. Initialize Repository

```bash
cd /path/to/joblet/homebrew-joblet
./scripts/init-repository.sh
```

**Expected outcomes:**
- [ ] Repository created with all files
- [ ] Initial commit pushed to main branch
- [ ] Directory structure in place
- [ ] GitHub issue templates configured

### 3. Configure Secrets

#### In homebrew repository (ehsaniara/homebrew-joblet):
- [ ] **GITHUB_TOKEN**: Automatically available (no action needed)

#### In main repository (ehsaniara/joblet):
- [ ] **HOMEBREW_UPDATE_TOKEN**: GitHub token with `repo` scope
  - Generate at: https://github.com/settings/tokens
  - Scopes needed: `repo`, `workflow`
  - Add to: https://github.com/ehsaniara/joblet/settings/secrets/actions

## 🧪 Pre-Deployment Testing

### 1. Formula Validation

```bash
cd homebrew-joblet
./scripts/test-formula.sh
```

**Must pass:**
- [ ] Formula syntax validation
- [ ] Content validation (all required fields)
- [ ] Installation options present
- [ ] Helper methods implemented

### 2. Main Repository Integration

- [ ] **Check release workflow**: Verify `quick-release.yml` includes homebrew trigger
- [ ] **Test admin UI build**: Ensure admin UI builds successfully
- [ ] **Verify archive structure**: Archives contain both `rnx` binary and `admin/` directory

### 3. Local Testing (if possible)

```bash
# Add tap locally for testing
brew tap ehsaniara/joblet /path/to/homebrew-joblet

# Test CLI-only installation
brew install rnx --without-admin
rnx --version
brew uninstall rnx

# Test admin installation (if Node.js available)
brew install rnx --with-admin
rnx --version
rnx-admin --help
brew uninstall rnx

# Clean up
brew untap ehsaniara/joblet
```

## 🚀 Deployment Process

### 1. Deploy to GitHub

If using the init script:
```bash
./scripts/init-repository.sh
```

Or manually:
```bash
git remote add origin https://github.com/ehsaniara/homebrew-joblet.git
git push -u origin main
```

### 2. Verify Repository

- [ ] **Repository is public**
- [ ] **All files are present**
- [ ] **Workflows are enabled**
- [ ] **Issues are enabled**

### 3. Test Auto-Update Workflow

#### Option A: Manual Workflow Trigger
1. Go to Actions tab in homebrew repository
2. Select "Update Formula" workflow
3. Click "Run workflow"
4. Provide test parameters:
   ```
   Version: v1.0.0-test
   Clean version: 1.0.0
   AMD64 URL: https://github.com/ehsaniara/joblet/releases/download/v1.0.0/rnx-v1.0.0-darwin-amd64.tar.gz
   ARM64 URL: https://github.com/ehsaniara/joblet/releases/download/v1.0.0/rnx-v1.0.0-darwin-arm64.tar.gz
   ```

#### Option B: End-to-End Test
1. Create a test release in main repository
2. Verify homebrew workflow triggers automatically
3. Check that formula updates with new checksums

**Expected outcomes:**
- [ ] Workflow completes successfully
- [ ] Formula updates with new URLs and checksums
- [ ] Commit is made with proper message
- [ ] No syntax errors in updated formula

## ✅ Post-Deployment Verification

### 1. User Installation Testing

```bash
# Test as a new user would
brew tap ehsaniara/joblet
brew install rnx

# Verify installation
rnx --version
rnx --help

# If admin UI installed
rnx admin --help
```

### 2. Comprehensive Verification

Use the verification script:
```bash
curl -fsSL https://raw.githubusercontent.com/ehsaniara/homebrew-joblet/main/scripts/verify-installation.sh | bash
```

**Expected results:**
- [ ] All critical checks pass
- [ ] No installation issues found
- [ ] Admin UI properly configured (if installed)
- [ ] CLI functionality works

### 3. Documentation Verification

- [ ] **README is accurate**: Installation commands work as documented
- [ ] **Links are valid**: All GitHub links point to correct repositories
- [ ] **Examples are current**: Version numbers and commands are up to date
- [ ] **Troubleshooting is complete**: Common issues are covered

## 🔧 Production Configuration

### 1. Branch Protection

Configure main branch protection:
- [ ] **Require pull request reviews**: For manual changes
- [ ] **Require status checks**: Ensure workflows pass
- [ ] **Restrict pushes**: Only maintainers can push directly
- [ ] **Allow force pushes**: Disabled for safety

### 2. Repository Settings

- [ ] **Issues enabled**: For user support
- [ ] **Discussions disabled**: Use issues instead
- [ ] **Wiki disabled**: Use docs/ directory
- [ ] **Sponsorship enabled**: If applicable
- [ ] **Security alerts enabled**: For dependency vulnerabilities

### 3. Monitoring Setup

- [ ] **Watch repository**: Enable notifications for issues
- [ ] **GitHub Actions notifications**: Enable failure notifications
- [ ] **Set up alerts**: For workflow failures

## 📚 User Communication

### 1. Announcement

Create announcement in main repository:

```markdown
🍺 **Homebrew Support Now Available!**

RNX is now available via Homebrew for macOS users:

```bash
brew tap ehsaniara/joblet
brew install rnx
```

Features:
- 🎯 Interactive installation with Node.js detection
- 🌐 Optional web admin UI
- 🚀 Auto-completions for bash, zsh, and fish
- 🔄 Automatic updates with new releases

For more details, see: https://github.com/ehsaniara/homebrew-joblet
```

### 2. Update Main Documentation

- [ ] **Add homebrew installation** to main README
- [ ] **Update INSTALLATION.md** with homebrew option
- [ ] **Add to quickstart guide**
- [ ] **Update CLI reference** with homebrew-specific notes

### 3. Community Communication

- [ ] **Update website** (if applicable)
- [ ] **Social media announcement** (if applicable)
- [ ] **Community forums** (if applicable)

## 🔍 Post-Launch Monitoring

### First 48 Hours

- [ ] **Monitor GitHub Issues**: Both repositories
- [ ] **Check workflow runs**: Ensure auto-updates work
- [ ] **Review installation reports**: From early adopters
- [ ] **Fix critical issues immediately**

### First Week

- [ ] **Gather user feedback**: Installation experience
- [ ] **Update troubleshooting docs**: Based on real issues
- [ ] **Performance check**: Installation times and success rates
- [ ] **Documentation improvements**: Based on user questions

### Ongoing

- [ ] **Regular testing**: Monthly installation tests
- [ ] **Dependency updates**: Keep Node.js requirements current
- [ ] **Security monitoring**: GitHub security alerts
- [ ] **Community engagement**: Respond to issues promptly

## 🚨 Rollback Plan

If critical issues are discovered:

### Immediate Response

1. **Disable auto-updates**:
   ```bash
   # In homebrew repository
   mv .github/workflows/update.yml .github/workflows/update.yml.disabled
   git add . && git commit -m "Temporarily disable auto-updates" && git push
   ```

2. **Revert to last known good version**:
   ```bash
   git log --oneline  # Find last good commit
   git revert <bad-commit-hash>
   git push origin main
   ```

3. **Communicate with users**:
   - Update README with known issues
   - Create GitHub issue with status update
   - Provide workaround instructions

### Recovery Process

1. **Fix the issue** in development environment
2. **Test thoroughly** with test script
3. **Deploy fix** via pull request
4. **Re-enable auto-updates** when stable
5. **Update community** on resolution

## 📊 Success Metrics

### Launch Week Goals

- [ ] **Zero critical failures**: No installations should fail completely
- [ ] **< 5 GitHub issues**: Minimal user-reported problems  
- [ ] **Documentation completeness**: Users can self-serve most issues
- [ ] **Auto-update success**: New releases trigger updates correctly

### Ongoing Health Metrics

- [ ] **Installation success rate**: > 95%
- [ ] **Issue resolution time**: < 24 hours for critical, < 72 hours for minor
- [ ] **User satisfaction**: Positive feedback on installation experience
- [ ] **Maintenance overhead**: < 1 hour/week for regular maintenance

## ✅ Final Deployment Checklist

Before marking deployment complete:

- [ ] ✅ Repository created and configured
- [ ] ✅ Secrets properly configured in both repositories  
- [ ] ✅ Formula passes all validation tests
- [ ] ✅ Auto-update workflow tested and working
- [ ] ✅ Manual installation tested end-to-end
- [ ] ✅ Documentation is accurate and complete
- [ ] ✅ Monitoring and alerts configured
- [ ] ✅ Rollback plan tested and ready
- [ ] ✅ User communication prepared
- [ ] ✅ Team trained on troubleshooting procedures

## 🎉 Post-Deployment

Once deployment is complete:

1. **Celebrate the milestone** 🎉
2. **Monitor for 48 hours** for any issues
3. **Gather initial user feedback**
4. **Plan next iteration** improvements
5. **Document lessons learned**

---

**Deployment Date**: _________________  
**Deployed By**: ___________________  
**Verified By**: ___________________  
**Status**: ⏳ Pending / ✅ Complete / ❌ Issues Found  

**Notes**: 
_________________________________
_________________________________
_________________________________