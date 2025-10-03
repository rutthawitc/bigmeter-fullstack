SELECT
    trn.CUST_CODE           AS "เลขที่ผู้ใช้น้ำ",
    tmp.METER_NO            AS "หมายเลขมาตร",
    cm.AVERAGE              AS "หน่วยน้ำเฉลี่ย",
    trn.PRESENT_METER_COUNT AS "เลขมาตรที่อ่านได้",
    trn.PRESENT_WATER_USG   AS "หน่วยน้ำปัจจุบัน",
    trn.DEBT_YM             AS "เดือนหนี้"
FROM
    PWACIS.TB_TR_DEBT_TRN trn
JOIN
    PWACIS.TB_TR_DEBT_TEMP tmp ON trn.CUST_ID = tmp.CUST_ID
                               AND trn.DEBT_YM = tmp.DEBT_YM
                               AND tmp.IS_DELETE = 'F'
LEFT JOIN
    PWACIS.TB_TR_CUSTOMER_INF cf ON trn.CUST_ID = cf.ID
LEFT JOIN
    PWACIS.TB_TR_CUST_METER cm ON cf.ID = cm.CUST_ID
                        AND cm.IS_DELETED = 'F'
LEFT JOIN
    PWACIS.TB_MS_METER_ROUTE mr ON tmp.METER_ROUTE_ID = mr.ID
LEFT JOIN
    PWACIS.TB_LT_METERSIZE ms ON tmp.METER_SIZE_ID = ms.ID
LEFT JOIN
    PWACIS.TB_LT_METERBRAND mb ON cm.MTR_BRAND_ID = mb.ID
LEFT JOIN
    PWACIS.TB_LT_METERSTATE mst ON cm.MRT_STATE_ID = mst.ID
LEFT JOIN
    PWACIS.TB_LT_USETYPE ut ON trn.CUST_TYPE_ID = ut.ID
WHERE
    trn.PRESENT_WATER_USG > 0
    AND trn.ORG_OWNER_ID = :ORG_OWNER_ID
    AND trn.DEBT_YM = :DEBT_YM
    AND trn.CANCELFLAG IS NULL
    AND trn.IS_DELETED = 'F'
    /*__CUSTCODE_FILTER__*/
ORDER BY
    trn.PRESENT_WATER_USG DESC
