package config

import ("errors";"os";"strings";"time")
type Config struct { HTTPAddr,CoreServiceToken,SynapseBaseURL,SynapseAdminToken string; RequestTimeout time.Duration }
func Load()(Config,error){ c:=Config{HTTPAddr:value("CORE_HTTP_ADDR",":8081"),CoreServiceToken:strings.TrimSpace(os.Getenv("CORE_SERVICE_TOKEN")),SynapseBaseURL:strings.TrimRight(strings.TrimSpace(os.Getenv("SYNAPSE_BASE_URL")),"/"),SynapseAdminToken:strings.TrimSpace(os.Getenv("SYNAPSE_ADMIN_TOKEN")),RequestTimeout:8*time.Second}; if c.CoreServiceToken==""||c.SynapseBaseURL==""||c.SynapseAdminToken=="" { return c,errors.New("CORE_SERVICE_TOKEN, SYNAPSE_BASE_URL and SYNAPSE_ADMIN_TOKEN are required") }; return c,nil }
func value(k,d string)string{if v:=strings.TrimSpace(os.Getenv(k));v!=""{return v};return d}
