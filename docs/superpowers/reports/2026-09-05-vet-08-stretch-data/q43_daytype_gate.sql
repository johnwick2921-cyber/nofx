-- Rec 6's own gate, applied to the 10 IB breaks in q40: at the fill instant the live
-- NY plan version must carry day_type=trend AND a bias matching the break side.
SELECT trade_date td, session, version v,
       datetime(strftime('%s',created_at),'unixepoch','-5 hours') born_ct,
       json_extract(doc,'$.day_type') day_type,
       COALESCE(json_extract(doc,'$.bias.direction'), json_extract(doc,'$.bias')) bias
  FROM plans WHERE session='NY'
   AND trade_date IN ('2026-08-20','2026-08-21','2026-08-25','2026-08-26','2026-08-27',
                      '2026-08-28','2026-09-01','2026-09-02','2026-09-03','2026-09-04')
 ORDER BY td, created_at;
-- day_type is FREE TEXT: ten distinct spellings across 254 plans (class 28).
SELECT json_extract(doc,'$.day_type') day_type, COUNT(*) n FROM plans GROUP BY 1 ORDER BY n DESC;
