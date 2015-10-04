# CMPE273-Assignment1

Author: Vrushank J Doshi
Date: 10/03/2015

Language: GO.

How to run Assignment 1?

Step 1: Run the server first using the following the command. Let the server be running.
        go run server.go
Step 2: Run the client application using the following command
        go run client.go
Step 3: Client application will ask for 3 options:
        Press 1 to buy stocks
        Press 2 to check your profile
        Press 3 to exit
Step 4: Choose your option. For eg. 1 and press Enter
Step 5: "Enter the stock symbol with percentage. For eg: GOOG:80,APPL:20"
        GOOG:70,APPL:30
        Note: Don't leave any space in between as it will throw you out and client terminates. If you provide 	any invalid stock symbol, it won't display any details.
Step 6: The request is sent to the server and server responds with following output:
        TradeID:1
        Stocks:<GOOG:10:$555>
        Unvested Amount:<$34>
Step 7: It will again ask the user for the following the options:
        Press 1 to buy stocks
        Press 2 to check your profile
        Press 3 to exit
Step 8: Now try selecting 2nd option. For eg: 2 and press Enter
Step 9: "Enter the trading ID to view your profile"
        1
Step 10:The trading ID is sent to the server and returns with the data associated with the trading ID.
        Stocks:<GOOG:10:$556>
        Current Market Value:<$5165>
        Unvested Amount:<$34>
Step 11:Choose the required option else press 3 to exit now to shut down the client and later shut down the 	server.

