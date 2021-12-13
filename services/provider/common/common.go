package common

const ProviderServicePath = "/provider.OCSProvider/"

func TokenAuthMethods() map[string]bool {
	return map[string]bool{
		ProviderServicePath + "OnBoardConsumer": true,
	}
}

func UIDAuthMethods() map[string]bool {
	return map[string]bool{
		ProviderServicePath + "GetStorageConfig": true,
		ProviderServicePath + "OffBoardConsumer": true,
		ProviderServicePath + "UpdateCapacity":   true,
	}
}
