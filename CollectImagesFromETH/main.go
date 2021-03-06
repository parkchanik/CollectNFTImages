package main

import (
	"bufio"
	"math/big"
	"strconv"

	"encoding/csv"

	"runtime"

	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	//"math"
	//"encoding/hex"
	"encoding/base64"
	"encoding/json"
	"errors"

	"bytes"
	//"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	//"github.com/ethereum/go-ethereum/core/types"
	//
	//"github.com/ethereum/go-ethereum/core/types"

	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"

	erc721 "CollectImagesFromETH/contracts/output/ERC721"

	//erc1155 "ETHCollectTrans/contracts/output/ERC1155"

	. "CollectImagesFromETH/types"

	logger "CollectImagesFromETH/logger"

	"CollectImagesFromETH/config"
	"net/http"
	"strings"
)

var client *ethclient.Client = nil

var IMAGE_PATH string = "./Images/"

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("GOMAXPROCS : ", runtime.GOMAXPROCS(0))

	configData, err := config.LoadConfigration("config.json")
	if err != nil {
		log.Fatal("LoadConfigration :", err)
	}

	url := configData.URL
	logger.LoggerInit()

	ethdial, err := ethclient.Dial(url)
	if err != nil {
		log.Fatal("ethclient.Dial ", err)
	}

	client = ethdial

	file, err := os.Open("./data/NFTNAMEFILEDOWN_20220205_2.csv")
	if err != nil {
		log.Fatal(err)
	}

	rdr := csv.NewReader(bufio.NewReader(file))

	rows, err := rdr.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range rows {

		//trx := row[1]
		contractAddress := row[0]
		//contractName := row[1]
		contractSymbol := row[1]
		tokenID := row[2]

		logger.InfoLog("-------------------------------------------------------contractAddressHex[%s] TokenID[%s]  ", contractAddress, tokenID)

		cAddress := common.HexToAddress(contractAddress)
		instance, err := erc721.NewErc721(cAddress, client)
		if err != nil {
			logger.InfoLog("Error NewErc721 contractAddressHex[%s] TokenID[%s] error[%s] ", contractAddress, tokenID, err.Error())
			continue
		}

		tokenIDInt, err := strconv.Atoi(tokenID)
		if err != nil {
			logger.InfoLog("Error strconv.Atoi contractAddressHex[%s] TokenID[%s] error[%s] ", contractAddress, tokenID, err.Error())
			continue
		}

		tokenIDBig := big.NewInt(int64(tokenIDInt))

		tokenURI, err := instance.TokenURI(&bind.CallOpts{}, tokenIDBig)
		if err != nil {
			logger.InfoLog("Error Token URI : contractAddressHex[%s] TokenID[%s] , error[%s] ", contractAddress, tokenID, err.Error())
			continue
		}

		fmt.Printf("%s ", tokenURI)
		fmt.Println()

		pathandfilename := fmt.Sprintf("%s%s_%s", IMAGE_PATH, contractSymbol, tokenID)

		tokenImagesFileName, subject := GetTokenURIData(tokenURI, pathandfilename)
		if err != nil {
			logger.InfoLog("Error getTokenMetaData : contractAddressHex[%s] TokenID[%s] , error[%s] ", contractAddress, tokenID, err.Error())
			continue
		}

		var b bytes.Buffer

		b.WriteString(contractAddress)
		b.WriteString(",")
		b.WriteString(contractSymbol)
		b.WriteString(",")
		b.WriteString(tokenID)
		b.WriteString(",")
		b.WriteString(subject)
		b.WriteString(",")
		b.WriteString(tokenImagesFileName)

		logger.ImageLog(b.String())

	}

}

func getTokenMetaData(tokenuri string) (TokenMetaData, error) {

	metadata := TokenMetaData{}

	//ipfs:// ??? ???????????? ??????????????? ??????
	// https://ipfs.io/ipfs/QmSTtv3w1jqcv5AKRRYVR5NN7fkTuuL9sNrkxRNL9e3fUo/4744 ???????????????

	// tokenuri
	//ipfs://QmWS694ViHvkTms9UkKqocv1kWDm2MTQqYEJeYi6LsJbxK ?????? ???????????????
	//ipfs://ipfs/QmWS694ViHvkTms9UkKqocv1kWDm2MTQqYEJeYi6LsJbxK ?????? ????????? ?????? ???????????????????????? ????????????
	// https://ipfs.io/ipfs/QmWS694ViHvkTms9UkKqocv1kWDm2MTQqYEJeYi6LsJbxK ????????? ?????????

	logger.InfoLog("-------tokenuri before : %s", tokenuri)

	r := strings.NewReplacer("ipfs://ipfs/", "https://ipfs.io/ipfs/", "ipfs://", "https://ipfs.io/ipfs/")

	tokenuri = r.Replace(tokenuri)

	logger.InfoLog("-------tokenuri after  %s", tokenuri)

	res, err := http.Get(tokenuri)
	if err != nil {
		logger.InfoLog("-------getTokenMetaData http.Get(tokenuri) tokenuri[%s] error[%s] ", tokenuri, err.Error())
		return metadata, err

	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.InfoLog("-------getTokenMetaData ioutil.ReadAll tokenuri[%s] error[%s] ", tokenuri, err.Error())
		return metadata, err

	}

	err = json.Unmarshal(data, &metadata)
	if err != nil {
		logger.InfoLog("-------getTokenMetaData  json.Unmarshal(data, &metadata)  data[%s] error[%s] ", string(data), err.Error())
		return metadata, err

	}

	return metadata, nil

}

// func getTokenImageUri(tokenuri string) (string, error) {

// 	//ipfs:// ??? ???????????? ??????????????? ??????

// 	// https://ipfs.io/ipfs/QmSTtv3w1jqcv5AKRRYVR5NN7fkTuuL9sNrkxRNL9e3fUo/4744 ???????????????

// 	if strings.Contains(tokenuri, "ipfs://") == true {

// 		tokenuri = strings.ReplaceAll(tokenuri, "ipfs://", "https://ipfs.io/ipfs/")

// 	}

// 	res, err := http.Get(tokenuri)
// 	if err != nil {
// 		return "", err
// 		//fmt.Printf("http Get Error Transaction[%s] , Tokenuri[%s] Error[%s]\n ", vLog.TxHash, tokenuri, err.Error())
// 	}

// 	defer res.Body.Close()

// 	data, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return "", err
// 		//fmt.Printf("res.Body error  Transaction[%s] , Tokenuri[%s] Error[%s]\n ", vLog.TxHash, tokenuri, err.Error())

// 	}
// 	metadata := TokenMetaData{}

// 	err = json.Unmarshal(data, &metadata)
// 	if err != nil {
// 		return "", err
// 		//fmt.Printf("metadata unmarshal error Transaction[%s] , Tokenuri[%s] Error[%s]\n ", vLog.TxHash, tokenuri, err.Error())

// 	}

// 	return metadata.Image, nil

// }

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url

	logger.InfoLog("start download image uri : %s , fileName : %s \n", URL, fileName)

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logger.InfoLog("-------downloadFile status code is not 200  URL[%s] fileName[%s] , code[%d]", URL, fileName, response.StatusCode)
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// func GetERC721Data(eventlog types.Log) (ContractAddr common.Address, Name string, Symbol string, TokenID string, err error) {

// 	err = nil

// 	ContractAddr = eventlog.Address
// 	Name = ""
// 	Symbol = ""
// 	TokenID = ""

// 	instance, err := erc721.NewErc721(eventlog.Address, client)
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 NewErc721 contractAddressHex[%s] , error[%s] ", ContractAddr.Hex(), err.Error())
// 		return
// 	}

// 	Name, err = instance.Name(&bind.CallOpts{})
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.Name error[%s] ", err.Error())

// 	}

// 	Symbol, err = instance.Symbol(&bind.CallOpts{})
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.Symbol error[%s] ", err.Error())

// 	}

// 	//0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef transfer
// 	erc721transfer, err := instance.ParseTransfer(eventlog)
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.ParseTransfer  error[%s] ", err.Error())
// 		return
// 	}

// 	TokenID = fmt.Sprintf("%s", erc721transfer.TokenId)

// 	logger.InfoLog("GetDataERC721  From[%s] , To[%s]  , TokenID[%d]", erc721transfer.From.Hex(), erc721transfer.To.Hex(), erc721transfer.TokenId.Int64())

// 	return

// }

// func GetDataERC721(eventlog types.Log) (ContractAddr string, Name string, Symbol string, TokenID string, TokenURI string, err error) {

// 	err = nil
// 	ContractAddr = ""
// 	Name = ""
// 	Symbol = ""
// 	TokenID = ""
// 	TokenURI = ""
// 	ContractAddr = eventlog.Address.Hex()

// 	instance, err := erc721.NewErc721(eventlog.Address, client)
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 NewErc721 contractAddressHex[%s] , error[%s] ", ContractAddr, err.Error())
// 		return
// 	}

// 	Name, err = instance.Name(&bind.CallOpts{})
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.Name error[%s] ", err.Error())
// 		return
// 	}

// 	Symbol, err = instance.Symbol(&bind.CallOpts{})
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.Symbol error[%s] ", err.Error())
// 		return
// 	}

// 	//0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef transfer
// 	erc721transfer, err := instance.ParseTransfer(eventlog)
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 instance.ParseTransfer  error[%s] ", err.Error())
// 		return
// 	}

// 	TokenID = fmt.Sprintf("%s", erc721transfer.TokenId)

// 	logger.InfoLog("GetDataERC721  From[%s] , To[%s]  , TokenID[%d]", erc721transfer.From.Hex(), erc721transfer.To.Hex(), erc721transfer.TokenId.Int64())

// 	TokenURI, err = instance.TokenURI(&bind.CallOpts{}, erc721transfer.TokenId)
// 	if err != nil {
// 		logger.InfoLog("GetDataERC721 Token URI : tokenid[%d] , error[%s] ", erc721transfer.TokenId.Int64(), err.Error())
// 		return
// 	}

// 	return

// }

func GetImageFromDataTextPlain(tokenuri, pathandfilename string) (rtn string, subject string) {

	rtn = ""
	subject = ""

	tokenuriarr := strings.Split(tokenuri, ",")

	tokenMetaData := TokenMetaDataBase64{}

	if strings.Trim(tokenuriarr[0], " ") == "data:text/plain;charset=utf-8" {

		// data := strings.Replace(tokenuri, "data:text/plain;charset=utf-8,", "", 1)

		// r, size := utf8.DecodeRune([]rune(data))
		// fmt.Printf("%c %v\n", r, size)

		// //logger.InfoLog("------- tokenuri uri [%s]\n", tokenuriarr[0])
		// err := json.Unmarshal([]byte(data), &tokenMetaData)
		// if err != nil {
		// 	logger.InfoLog("GetImageFromDataTextPlain tokenMetaData utf8 Unmarshal Error : ", err)
		// 	logger.InfoLog("token string [%s]\n", tokenuriarr[1])
		// 	return
		// }

	}

	//imagearr := strings.Split(tokenMetaData.Image, ",")

	subject = tokenMetaData.Name

	// file, err := os.Create(pathandfilename)
	// if err != nil {
	// 	logger.InfoLog("getImageFromDataApplicationJson os.Create Error : ", err)
	// 	return
	// }

	// defer file.Close()

	return

}

func GetImageFromDataApplicationJson(tokenuri, pathandfilename string) (rtn string, subject string) {

	logger.InfoLog("------- tokenuri uri [%s]\n", "data:application/json........")

	//logger.InfoLog("token uri data:json : imageuri uri %s\n", tokenuri)

	tokenuriarr := strings.Split(tokenuri, ",")

	tokenMetaData := TokenMetaDataBase64{}

	rtn = ""
	subject = ""

	if strings.Trim(tokenuriarr[0], " ") == "data:application/json;utf8" {

		//logger.InfoLog("token uri data:json : strings.Replace(tokenuri, data:application/json;utf8 uri %s\n", strings.Replace(tokenuri, "data:application/json;utf8,", "", 1))

		data := strings.Replace(tokenuri, "data:application/json;utf8,", "", 1)

		//logger.InfoLog("------- tokenuri uri [%s]\n", tokenuriarr[0])
		err := json.Unmarshal([]byte(data), &tokenMetaData)
		if err != nil {
			logger.InfoLog(" tokenMetaData utf8 Unmarshal Error : ", err)
			logger.InfoLog("token string [%s]\n", tokenuriarr[1])
			return
		}

	} else if strings.Trim(tokenuriarr[0], " ") == "data:application/json;base64" {

		logger.InfoLog("------- tokenuri uri [%s]\n", tokenuriarr[0])

		data, err := base64.StdEncoding.DecodeString(tokenuriarr[1])
		if err != nil {
			logger.InfoLog(" tokenMetaData base64.StdEncoding.DecodeString Error : ", err)
			return
		}

		//fmt.Printf("test data : %s\n", string(data))

		err = json.Unmarshal(data, &tokenMetaData)
		if err != nil {
			logger.InfoLog(" tokenMetaData base64 Unmarshal Error : ", err)
			logger.InfoLog("token DecodeString [%s]\n", string(data))
			return
		}

	} else {

		logger.InfoLog("------- tokenuri uri not  data:application/json;utf8 and  data:application/json;base64 [%s]\n", tokenuriarr[0])
		return
	}

	//logger.InfoLog("token uri data:json : imageuri tokenuriarr[1]  ---- uri [%s]\n", tokenuriarr[1])

	imagearr := strings.Split(tokenMetaData.Image, ",")

	subject = tokenMetaData.Name

	file, err := os.Create(pathandfilename)
	if err != nil {
		logger.InfoLog("getImageFromDataApplicationJson os.Create Error : ", err)
		return
	}

	defer file.Close()

	//logger.InfoLog("tokenMetaData.Image[%s]\n", tokenMetaData.Image)

	if strings.Trim(imagearr[0], " ") == "data:image/svg+xml;utf8" {

		//logger.InfoLog("data:image/svg+xml;utf8 imagearr[1][%s]\n", imagearr[2])

		imageUTF8 := strings.Replace(tokenMetaData.Image, "data:image/svg+xml;utf8,", "", 1)

		cnt, err := file.WriteString(imageUTF8)
		if err != nil {
			logger.InfoLog("getImageFromDataApplicationJson data:image/svg+xml;utf8 file.WriteString Error : ", err)
			return
		}

		logger.InfoLog("file.WriteString data:image/svg+xml;utf8 cnt %d ", cnt)

		rtn = "OK"
		return

	} else if strings.Trim(imagearr[0], " ") == "data:image/svg+xml;base64" { // svg , base64 ??? ????????? ???????????? ?????? svg ??? ?????????
		imgdata, err := base64.StdEncoding.DecodeString(imagearr[1])
		if err != nil {
			logger.InfoLog("base64.StdEncoding.DecodeString(imagearr Error : ", err)
			return
		}

		//logger.InfoLog("base64.StdEncoding.DecodeString  %s\n", imgdata)

		cnt, err := file.WriteString(string(imgdata))
		if err != nil {
			logger.InfoLog("getImageFromDataApplicationJson data:image/svg+xml;base64 file.WriteString Error : ", err)
			return
		}

		logger.InfoLog("file.WriteString data:image/svg+xml;base64 cnt %d ", cnt)
		rtn = "OK"
		return
	}

	return
}

func GetTokenURIData(tokenuri, pathandfilename string) (rtn string, subject string) {

	//replacer := strings.NewReplacer(" ", "_", ":", "", "?", "", "*", "", "<", "", ">", "", "|", "", "\"", "", "/", "")
	//contractNameFilter := replacer.Replace(contractName)

	rtn = ""
	subject = ""
	if strings.Contains(tokenuri, "data:application/json") == true {

		pathandfilename := fmt.Sprintf("%s.svg", pathandfilename)
		result, name := GetImageFromDataApplicationJson(tokenuri, pathandfilename)

		rtn = pathandfilename
		subject = name
		if result == "OK" {
			rtn = "OK"
			subject = name

		} else {
			logger.InfoLog("GetImageFromDataApplicationJson Result Not OK Tokenuri[%s] , FileName[%s] \n ", tokenuri, pathandfilename)
		}

	} else if strings.Contains(tokenuri, "data:text/plain;charset=utf-8") == true {

		result, name := GetImageFromDataTextPlain(tokenuri, pathandfilename)

		rtn = pathandfilename
		subject = name
		if result == "OK" {
			rtn = "OK"
			subject = name

		} else {
			logger.InfoLog("GetImageFromDataTextPlain Result Not OK Tokenuri[%s] , FileName[%s] \n ", tokenuri, pathandfilename)
		}

	} else {

		logger.InfoLog("------- tokenuri uri [%s]\n", tokenuri)

		tokenMetaData, err := getTokenMetaData(tokenuri)
		if err != nil {
			logger.InfoLog("--------------------------getTokenImageUri , Tokenuri[%s] Error[%s]\n ", tokenuri, err.Error())
		} else {

			imageuri := tokenMetaData.Image

			subject = fmt.Sprintf("%v", tokenMetaData.Name)

			pathandfilename := fmt.Sprintf("%s.png", pathandfilename)

			rtn = pathandfilename

			if strings.Contains(imageuri, "ipfs://") == true {
				imageuri = strings.ReplaceAll(imageuri, "ipfs://", "https://ipfs.io/ipfs/")
			}

			if strings.Contains(imageuri, "ipfs") == true { /// 20220116 ipfs ?????? image ??????????????? ?????? ???????????????  ?????? ????????? ?????? download pas

				logger.InfoLog("------ipfs image url!! Tokenuri[%s] FileName[%s] ,  ImageURL[%s]\n ", tokenuri, pathandfilename, imageuri)
			} else {

				/*
					err = downloadFile(imageuri, pathandfilename)
					if err != nil {
						logger.InfoLog("--------------------------downloadfile error Transaction[%s] , Image[%s] , FileName[%s] , Error[%s]\n ", imageuri, pathandfilename, err.Error())

					}
				*/
			}
		}

	}

	return

}
