package main

func Example() {
	main()
	// Output:
	// L=1 checkout: order created
	// L=2 checkout: sent payment request
	// L=3 payment: received payment request
	// L=4 payment: charged card
	// L=5 checkout: order confirmed
}
