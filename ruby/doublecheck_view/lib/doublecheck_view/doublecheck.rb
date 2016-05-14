module DoublecheckView
  class ViewResult
    attr_reader :name, :rows

    def initialize(name, rows)
      @name = name
      @rows = rows
    end

    def valid?
      rows.empty?
    end
  end

  class CheckResult
    attr_reader :view_results

    def initialize(view_results)
      @view_results = view_results
    end

    def valid?
      view_results.all?(&:valid?)
    end

    def errors
      view_results.reject(&:valid?).each_with_object({}) do |vr, hash|
        hash[vr.name] = vr.rows
      end
    end
  end

  class Doublecheck
    attr_reader :schema
    attr_reader :views

    def initialize(schema: "doublecheck")
      @schema = schema
      @views = get_views
    end

    def check(views: self.views)
      view_results = views.map do |v|
        rows = conn.select_all("select * from #{quote_identifier schema}.#{quote_identifier v}").to_a
        ViewResult.new v, rows
      end

      CheckResult.new view_results
    end

    private

    def get_views
      conn.select_values <<-SQL
        select table_name from information_schema.views where table_schema=#{quote_value schema} order by 1
      SQL
    end

    def conn
      ActiveRecord::Base.connection
    end

    def quote_value(value)
      conn.quote value
    end

    def quote_identifier(ident)
      conn.quote_column_name ident
    end
  end
end
