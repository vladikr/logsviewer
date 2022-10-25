export type Pod = {
	uuid: string;
	name: string;
	namespace: string;
	phase: string;
	activeContainers: number;
	totalContainers: number;
	creationTime: Date;
};

export type PodsApiResponse = {
    data: Pod[];
    meta: {
  	    totalRowCount: number;
    }
};
