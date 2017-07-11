main:                                   # @main
        mov     al, 12
        jmp     .LBB0_1
.LBB0_6:                                #   in Loop: Header=BB0_1 Depth=1
        add     al, cl
        movzx   ecx, al
        shl     rcx, 5
        mov     byte ptr [rcx + 22530], 32
        mov     byte ptr [22942], 16
.LBB0_1:                                # =>This Inner Loop Header: Depth=1
        movzx   edx, byte ptr [0]
        test    dl, 8
        jne     .LBB0_2
        and     dl, 4
        shr     dl, 2
        cmp     al, 2
        seta    cl
        and     cl, dl
        mov     dl, -2
        test    cl, cl
        je      .LBB0_6
        jmp     .LBB0_5
.LBB0_2:                                #   in Loop: Header=BB0_1 Depth=1
        cmp     al, 22
        setb    cl
        mov     dl, 2
        test    cl, cl
        je      .LBB0_6
.LBB0_5:                                #   in Loop: Header=BB0_1 Depth=1
        mov     ecx, edx
        jmp     .LBB0_6
