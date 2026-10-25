# Project purpose

These are small learning projects for understanding distributed systems and operating systems more deeply.

- Keep implementations minimal and easy to follow, rather than production oriented.
- Make only changes needed for the concept being learned.
- Preserve the concept's essential mechanisms; consistent hashing includes virtual nodes.
- For every library, design from the caller's perspective: who uses it, what they build, and how they integrate it. A library is incomplete without a small runnable consumer example demonstrating a realistic use case, commands to run it, and the expected behavior. Keep the example minimal and verify it works.
- Keep only necessary comments. Avoid extra abstractions, dependencies, and configuration.
- Reuse each project's existing directory, including `consistenthash` for the Go consistent hashing project.
- Use `mijun/<feature-name>` for feature branches.
- Keep commits and pull requests focused on the technical change, without agent attribution.
