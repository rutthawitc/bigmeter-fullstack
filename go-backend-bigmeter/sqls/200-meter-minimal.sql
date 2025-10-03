-- Optimized: filter + deduplicate first, then join heavy dimensions after limiting to 200 rows
WITH base AS (
    SELECT /*+ MATERIALIZE */
        trn.CUST_ID,
        trn.ORG_OWNER_ID AS BA,
        trn.CUST_CODE,
        trn.CUST_TYPE_ID,
        trn.CUST_NAME,
        trn.CUST_ADDRESS,
        tmp.METER_ROUTE_ID,
        tmp.METER_SIZE_ID,
        tmp.METER_NO,
        trn.DEBT_YM,
        trn.PRESENT_WATER_USG
    FROM PWACIS.TB_TR_DEBT_TRN trn
    JOIN PWACIS.TB_TR_DEBT_TEMP tmp
      ON trn.CUST_ID = tmp.CUST_ID
     AND trn.DEBT_YM = tmp.DEBT_YM
     AND tmp.IS_DELETE = 'F'
    WHERE trn.PRESENT_WATER_USG > 0
      AND trn.ORG_OWNER_ID = :ORG_OWNER_ID
      AND trn.DEBT_YM = :DEBT_YM
      AND trn.CANCELFLAG IS NULL
      AND trn.IS_DELETED = 'F'
), dedup AS (
    SELECT /*+ MATERIALIZE */
        b.*,
        ROW_NUMBER() OVER (PARTITION BY b.CUST_CODE ORDER BY b.PRESENT_WATER_USG DESC) AS rn
    FROM base b
), top200 AS (
    SELECT /*+ MATERIALIZE */
        d.*
    FROM dedup d
    WHERE d.rn = 1
    ORDER BY d.PRESENT_WATER_USG DESC
    FETCH FIRST 200 ROWS ONLY
)
SELECT
    t.BA                           AS "BA",
    org.org_name                   AS "แม่ข่าย/หน่วยบริการ",
    t.CUST_CODE                    AS "เลขที่ผู้ใช้น้ำ",
    ut.USETYPE                     AS "ประเภท",
    ut.usename                     AS "รายละเอียด",
    t.CUST_NAME                    AS "ชื่อผู้ใช้น้ำ",
    t.CUST_ADDRESS                 AS "ที่อยู่",
    mr.METER_ROUTE_CODE            AS "เส้นทาง",
    t.METER_NO                     AS "หมายเลขมาตร",
    ms.SIZENAME                    AS "ขนาดมาตร",
    mb.BRANDNAME                   AS "ยี่ห้อมาตร",
    mst.STATENAME                  AS "สถานะมาตร",
    t.DEBT_YM                      AS "เดือนหนี้"
FROM top200 t
LEFT JOIN PWACIS.TB_TR_CUST_METER cm ON t.CUST_ID = cm.CUST_ID AND cm.IS_DELETED = 'F'
LEFT JOIN PWACIS.TB_LT_METERSTATE mst ON cm.MRT_STATE_ID = mst.ID
LEFT JOIN PWACIS.TB_MS_METER_ROUTE mr ON t.METER_ROUTE_ID = mr.ID
LEFT JOIN PWACIS.TB_MS_METER_LINE ml ON mr.METER_LINE_ID = ml.ID
LEFT JOIN PWACIS.TB_LT_ORGANIZATION org ON ml.ORG_CC_ID = org.ID
LEFT JOIN PWACIS.TB_LT_METERSIZE ms ON t.METER_SIZE_ID = ms.ID
LEFT JOIN PWACIS.TB_LT_METERBRAND mb ON cm.MTR_BRAND_ID = mb.ID
LEFT JOIN PWACIS.TB_LT_USETYPE ut ON t.CUST_TYPE_ID = ut.ID
ORDER BY t.PRESENT_WATER_USG DESC
