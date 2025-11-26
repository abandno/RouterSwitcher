export namespace main {
	
	export class Config {
	    HomeSSID: string;
	    StaticIP: string;
	    Gateway: string;
	    DNS: string;
	    AutoStart: boolean;
	    IPMode: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.HomeSSID = source["HomeSSID"];
	        this.StaticIP = source["StaticIP"];
	        this.Gateway = source["Gateway"];
	        this.DNS = source["DNS"];
	        this.AutoStart = source["AutoStart"];
	        this.IPMode = source["IPMode"];
	    }
	}

}

