// Load the files (assuming you copied ABI and BIN)
var abi = [{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"set","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"storedValue","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]
var bytecode = "0x" + "608060405234801561001057600080fd5b5060b18061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c806360fe47b11460375780636d619daa146049575b600080fd5b604760423660046063565b600055565b005b605160005481565b60405190815260200160405180910390f35b600060208284031215607457600080fd5b503591905056fea2646970667358221220c9f276b4b7387497ae4dfedb9cafc166f33134020bb63e19a8fb5dcc09290fe364736f6c63430008180033";

// Define accounts
var from = ftm.accounts[0];
var to = ftm.accounts[1];

// Unlock the sender (use your actual password)
personal.unlockAccount(from, "fakepassword", 60);

// Send transaction
var txHash = ftm.sendTransaction({
    from: from,
    to: to,
    value: web3.toWei(1, "ether")  // send 1 ETH, adjust as needed
});

txHash


// Unlock your account to deploy
personal.unlockAccount(personal.listAccounts[1], "hello", 0)

// Create contract object
var SimpleStorage = ftm.contract(abi);

// Deploy
var simpleStorage = SimpleStorage.new({
    from: ftm.accounts[1],
    data: bytecode,
    gas: 3000000
}, function(err, contract) {
    if (!err) {
        if (!contract.address) {
            console.log("Tx hash: " + contract.transactionHash);
        } else {
            console.log("Contract deployed at: " + contract.address);
        }
    }
});
