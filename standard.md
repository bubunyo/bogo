# bogo standard

1st Byte - |_|_|_|_|_|_|_| - encoding version, (Allows for only 255 version)
next 1 byte - type
selct type {
    type: object
        - next byte space held by key
        - next 4 bytes - size of key, lets assign a value of kn. 
        - next 4 bytes - size of value - lets assign a value of vn
        - next kn bytes - key
        - next vn bytes - value 
    type: bool:
        # tbd
    type: array:
        # tbd
    type: null:
        # tbd
    type: number:
        - byte 1 -
            - bit 1 - sign
            - bit 2 - 7 - space required. Number of bytes in powers of 2. lets this be n
        - byte 2 - 1 + 2^n - data
            - encoding is little endian
            
    
}
  

