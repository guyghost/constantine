## Description
<!-- Provide a brief description of the changes in this PR -->

## Type of Change
<!-- Mark the relevant option with an 'x' -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Code style update (formatting, renaming)
- [ ] ♻️ Code refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] ✅ Test update
- [ ] 🔧 Build configuration change
- [ ] 🚀 CI/CD update
- [ ] 🔒 Security fix

## Related Issues
<!-- Link to related issues using #issue_number -->

Fixes #
Relates to #

## Changes Made
<!-- List the main changes made in this PR -->

- 
- 
- 

## Testing
<!-- Describe the tests you ran and how to reproduce them -->

- [ ] All existing tests pass (`make test-race`)
- [ ] New tests added for new functionality
- [ ] Manual testing performed
- [ ] Code coverage maintained or improved

**Test commands run:**
```bash
make test-race
make lint
make ci
```

## Security Considerations
<!-- Describe any security implications of this change -->

- [ ] No sensitive data exposed
- [ ] No new dependencies added OR dependencies scanned with `make vulncheck`
- [ ] Security scan passed (`make audit`)
- [ ] No hardcoded secrets or credentials

## Documentation
<!-- Check all that apply -->

- [ ] Code comments added/updated
- [ ] README updated (if needed)
- [ ] Documentation in `/docs` updated (if needed)
- [ ] API documentation updated (if applicable)

## Checklist
<!-- Ensure all items are checked before requesting review -->

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
- [ ] I have run `make ci` locally and all checks pass

## Performance Impact
<!-- Describe any performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance may be affected (describe below)

**Details:**


## Breaking Changes
<!-- If this is a breaking change, describe what breaks and migration steps -->

**What breaks:**


**Migration guide:**


## Screenshots/Logs
<!-- If applicable, add screenshots or logs to help explain your changes -->


## Additional Notes
<!-- Add any other context about the PR here -->


---

**Reviewer Guidelines:**
- Verify all CI checks pass
- Check code coverage hasn't decreased
- Review security implications
- Ensure documentation is updated
- Test manually if needed
