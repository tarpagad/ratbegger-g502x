export namespace main {
	
	export class Device {
	    codename: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.codename = source["codename"];
	        this.name = source["name"];
	    }
	}

}

