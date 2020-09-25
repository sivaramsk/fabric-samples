/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//var logger = shim.NewLogger("userrolecc")

// SimpleChaincode example simple Chaincode implementation
type UserRole struct {
	UserRoleID   string `json:"UserRoleID"`
	UserRoleName string `json:"UserRoleName"`
}

func (t *UserRole) Init(stub shim.ChaincodeStubInterface) pb.Response {
//	logger.Info("########### UserRole Init ###########")

	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *UserRole) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
//	logger.Info("########### UserRole Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "CreateUserRole" {
		// Deletes an entity from its state
		return t.CreateUserRole(stub, args)
	}

	if function == "GetUserRole" {
		// queries an entity state
		return t.GetUserRole(stub, args)
	}

	if function == "GetUserByRoleName" {
		// queries an entity state
		return t.GetUserByRoleName(stub, args)
	}

	if function == "GetAllUserRoles" {
		// queries an entity state
		return t.GetAllUserRoles(stub, args)
	}

//	logger.Errorf("Unknown action, check the first argument, must be one of 'CreateUserRole', 'GetUserRole' or 'GetAllUserRoles' or 'GetUserByRoleName'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'CreateUserRole', 'GetUserRole' or 'GetAllUserRoles' or 'GetUserByRoleName' . But got: %v", args[0]))
}

func (t *UserRole) CreateUserRole(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) < 2 {
		return shim.Error("Incorrect arguments. Expecting 2 arguments")
	}
	// json args input musts match  the case and  spelling exactly
	// get all the  arguments
	var userRoleId = args[0]
	fmt.Println("the UserRoleID is" + userRoleId)
	var userRoleName = args[1]
	fmt.Println("The UserRoleName is" + userRoleName)

	//assigning to struct the variables
	userRoleStruct := UserRole{
		UserRoleID:   userRoleId,
		UserRoleName: userRoleName,
	}

	userRoleStructBytes, err := json.Marshal(userRoleStruct)
	if err != nil {
		return shim.Error("Unable to unmarshal data from struct")
	}

	err = stub.PutState(userRoleId, []byte(userRoleStructBytes))
	if err != nil {
		return shim.Error(err.Error())
	}

	//creating Composite Key index to be able to query based on it
	indexName := "roleIdNameIndex"
	roleIDNameKey, err := stub.CreateCompositeKey(indexName, []string{userRoleStruct.UserRoleName, userRoleStruct.UserRoleID})
	if err != nil {
		fmt.Println("Could not index composite key for RoleId and RoleName", err)
		return shim.Error("Could not index composite key for RoleId and Name")
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the collection.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(roleIDNameKey, value)

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *UserRole) GetUserRole(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var userRoleId string // userRoleId
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting userRoleId to query")
	}

	userRoleId = args[0]

	// Get the state from the ledger
	userRoleBytes, err := stub.GetState(userRoleId)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + userRoleId + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"User Role\":\"" + string(userRoleBytes) + "\"}"
        fmt.Println(jsonResp)
//	logger.Infof("Query Response:%s\n", jsonResp)
	return shim.Success(userRoleBytes)
}

// Query callback representing the query of a chaincode
func (t *UserRole) GetUserByRoleName(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		fmt.Println("Invalid number of arguments")
		return shim.Error("Missing Arguments")
	}

	roleName := args[0]

	GetRoleIDNameItr, err := stub.GetStateByPartialCompositeKey("roleIdNameIndex", []string{roleName})
	if err != nil {
		return shim.Error("Could not get Role")
	}
	defer GetRoleIDNameItr.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer

	bArrayMemberAlreadyWritten := false
	for GetRoleIDNameItr.HasNext() {
		queryResponse, err := GetRoleIDNameItr.Next()
		if err != nil {
			return shim.Error("Could not get User Role Data")
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(queryResponse.Key)
		if err != nil {
			return shim.Error("Could not get Composite Key Parts")
		}
		returnedRoleName := compositeKeyParts[0]
		returnedRoleID := compositeKeyParts[1]

		fmt.Printf("- found a  from index:%s RoleName:%s RoleID:%s\n", objectType, returnedRoleName, returnedRoleID)

		value, err := stub.GetState(returnedRoleID)
		if err != nil {
			fmt.Println("Could not get details for id "+returnedRoleID+" from ledger", err)
			return shim.Error("Missing RoleID")
		}
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(value))
		bArrayMemberAlreadyWritten = true
	}

	return shim.Success(buffer.Bytes())
}

// Query callback representing the query of a chaincode
func (t *UserRole) GetAllUserRoles(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error

	//Get all Document Groups from the ledger
	GetAllUserRolesItr, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error("Could not get all User Roles")
	}
	defer GetAllUserRolesItr.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer

	// buffer.WriteString("{\"Results\":")

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for GetAllUserRolesItr.HasNext() {
		queryResponseValue, err := GetAllUserRolesItr.Next()
		if err != nil {
			return shim.Error("Could not get next User Role")
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponseValue.Value))

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	// buffer.WriteString("}")

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(UserRole))
        {
		fmt.Printf("Error starting UserRole chaincode: %s", err)
	}
}
