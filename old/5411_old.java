            response.setErrorInfo(NacosException.SERVER_ERROR, "Nacos cluster is running with 1.X mode, can't accept gRPC request temporarily. Please check the server status or close Double write to force open 2.0 mode. Detail https://nacos.io/en-us/docs/2.0.0-upgrading.html.");
            return response;
        }
        return null;
    }
}
