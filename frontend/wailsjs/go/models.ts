export namespace appinfo {
	
	export class Info {
	    version: string;
	    commit: string;
	    platform: string;
	    data_dir: string;
	    log_file: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.platform = source["platform"];
	        this.data_dir = source["data_dir"];
	        this.log_file = source["log_file"];
	    }
	}

}

