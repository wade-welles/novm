package plan9

func checkBuffer(buf Buffer) error {

	// Did we overrun?
	if buf.WriteLeft() < 0 {
		return BufferInsufficient
	}

	return nil
}

func packCommon(buf Buffer, size int, id uint8, tag uint16) error {
	buf.WriteRewind()
	size += 4 + 1 + 2 // size[4] id[1] tag[2]
	if buf.WriteLeft() < int(size) {
		return BufferInsufficient
	}

	buf.Write32(uint32(size))
	buf.Write8(id)
	buf.Write16(tag)

	return nil
}

func PackRversion(buf Buffer, tag uint16, msize uint32, version string) error {
	size := 4 + 2 + len(version) // msize[4] version[s]
	err := packCommon(buf, size, Rversion, tag)
	if err != nil {
		return err
	}
	buf.Write32(msize)
	buf.WriteString(version)

	return checkBuffer(buf)
}

func PackRauth(buf Buffer, tag uint16, aqid *Qid) error {
	size := 13 // aqid[13]
	err := packCommon(buf, size, Rauth, tag)
	if err != nil {
		return err
	}

	pqid(buf, aqid)

	return checkBuffer(buf)
}

func PackRerror(buf Buffer, tag uint16, errString string, errornum uint32, dotu bool) error {
	size := 2 + len(errString) // ename[s]
	if dotu {
		size += 4 // ecode[4]
	}
	err := packCommon(buf, size, Rerror, tag)
	if err != nil {
		return err
	}
	buf.WriteString(errString)
	if dotu {
		buf.Write32(errornum)
	}

	return checkBuffer(buf)
}

func PackRflush(buf Buffer, tag uint16) error { return packCommon(buf, 0, Rflush, tag) }
func PackRattach(buf Buffer, tag uint16, aqid *Qid) error {
	size := 13 // aqid[13]
	err := packCommon(buf, size, Rattach, tag)
	if err != nil {
		return err
	}

	pqid(buf, aqid)

	return checkBuffer(buf)
}

func PackRwalk(buf Buffer, tag uint16, wqids []Qid) error {
	nwqid := len(wqids)
	size := 2 + nwqid*13 // nwqid[2] nwname*wqid[13]
	err := packCommon(buf, size, Rwalk, tag)
	if err != nil {
		return err
	}

	buf.Write16(uint16(nwqid))
	for i := 0; i < nwqid; i++ {
		pqid(buf, &wqids[i])
	}

	return checkBuffer(buf)
}

func PackRopen(
	buf Buffer,
	tag uint16,
	qid *Qid,
	iounit uint32) error {

	size := 13 + 4 // qid[13] iounit[4]
	err := packCommon(buf, size, Ropen, tag)
	if err != nil {
		return err
	}

	pqid(buf, qid)
	buf.Write32(iounit)

	return checkBuffer(buf)
}

func PackRcreate(
	buf Buffer,
	tag uint16,
	qid *Qid,
	iounit uint32) error {

	size := 13 + 4 // qid[13] iounit[4]
	err := packCommon(buf, size, Rcreate, tag)
	if err != nil {
		return err
	}

	pqid(buf, qid)
	buf.Write32(iounit)

	return checkBuffer(buf)
}

func PackRread(
	buf Buffer,
	tag uint16,
	count uint32) error {

	size := int(4 + count) // count[4] data[count]
	err := packCommon(buf, size, Rread, tag)
	if err != nil {
		return err
	}

	buf.Write32(count)

	return checkBuffer(buf)
}

func PackRwrite(
	buf Buffer,
	tag uint16,
	count uint32) error {

	err := packCommon(buf, 4, Rwrite, tag) // count[4]
	if err != nil {
		return err
	}

	buf.Write32(count)
	return checkBuffer(buf)
}

func PackRclunk(
	buf Buffer,
	tag uint16) error {

	return packCommon(buf, 0, Rclunk, tag)
}

func PackRremove(buf Buffer, tag uint16) error { return packCommon(buf, 0, Rremove, tag) }

func PackRstat(buf Buffer, tag uint16, d *Dir, dotu bool) error {
	stsz := statsz(d, dotu)
	size := 2 + stsz // stat[n]
	err := packCommon(buf, size, Rstat, tag)
	if err != nil {
		return err
	}

	buf.Write16(uint16(stsz))
	pstat(buf, d, dotu)
	return checkBuffer(buf)
}

func PackRwstat(buf Buffer, tag uint16) error { return packCommon(buf, 0, Rwstat, tag) }
