package main

func Example() {
	main()
	// Output:
	// "Add diagram" vs "Add tests": concurrent
	// Keep both revisions for a user to merge.
	// "Add diagram" vs merged: before
	// "Add tests" vs merged: before
	// Merged revision: "Add diagram and tests"
}
