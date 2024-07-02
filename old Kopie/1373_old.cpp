/******************************************************************************
  auto negate_pred = [](predicate p) { return negation{expression{p}}; };
  auto negate_expr = [](expression expr) { return negation{expr}; };
 *                    _   _____   __________                                  *
 *                   | | / / _ | / __/_  __/     Visibility                   *
 *                   | |/ / __ |_\ \  / /          Across                     *
 *                   |___/_/ |_/___/ /_/       Space and Time                 *
 *                                                                            *
 * This file is part of VAST. It is subject to the license terms in the       *
 * LICENSE file found in the top-level directory of this distribution and at  *
 * http://vast.io/license. No part of VAST, including this file, may be       *
 * copied, modified, propagated, or distributed except according to the terms *
 * contained in the LICENSE file.                                             *
 ******************************************************************************/

#include "vast/concept/parseable/vast/expression.hpp"

#include "vast/concept/parseable/core/parser.hpp"
#include "vast/concept/parseable/vast/data.hpp"
#include "vast/concept/parseable/vast/type.hpp"
#include "vast/data.hpp"
#include "vast/detail/assert.hpp"
#include "vast/detail/string.hpp"
#include "vast/expression.hpp"
#include "vast/expression_visitors.hpp"
#include "vast/type.hpp"

namespace vast {

namespace {

using predicate_tuple
  = std::tuple<predicate::operand, relational_operator, predicate::operand>;

static predicate to_predicate(predicate_tuple xs) {
  return {std::move(std::get<0>(xs)), std::get<1>(xs),
          std::move(std::get<2>(xs))};
}

static predicate::operand to_field_extractor(std::vector<std::string> xs) {
  // TODO: rather than doing all the work with the attributes, it would be nice
  // if the parser framework would just give us an iterator range over the raw
  // input. Then we wouldn't have to use expensive attributes and could simply
  // wrap a parser P into raw(P) or raw_str(P) to obtain a range/string_view.
  auto field = detail::join(xs, ".");
  return field_extractor{std::move(field)};
}

static predicate::operand to_attr_extractor(std::string x) {
  return attribute_extractor{caf::atom_from_string(x)};
}

static predicate::operand to_type_extractor(type x) {
  return type_extractor{std::move(x)};
}

static predicate::operand to_data_operand(data x) {
  return x;
}

static predicate to_value_predicate(data x) {
  auto infer_type = [](auto& d) -> type {
    return data_to_type<std::decay_t<decltype(d)>>{};
  };
  auto lhs = type_extractor{caf::visit(infer_type, x)};
  auto rhs = predicate::operand{std::move(x)};
  return {std::move(lhs), relational_operator::equal, std::move(rhs)};
}

static auto make_predicate_parser() {
  using parsers::alnum;
  using parsers::chr;
  using namespace parser_literals;
  // clang-format off
  auto id = +(alnum | chr{'_'} | chr{'-'});
  auto field_char = alnum | chr{'_'} | chr{'-'} | chr{':'};
  // A field cannot start with:
  //  - '-' to leave room for potential arithmetic expressions in operands
  //  - ':' so it won't be interpreted as a type extractor
  auto field = !(':'_p | '-') >> (+field_char % '.');
  auto operand
    = (parsers::data >> !(field_char | '.')) ->* to_data_operand
    | '#' >> id ->* to_attr_extractor
    | ':' >> parsers::type ->* to_type_extractor
    | field ->* to_field_extractor
    ;
  auto operation
    = "~"_p   ->* [] { return relational_operator::match; }
    | "!~"_p  ->* [] { return relational_operator::not_match; }
    | "=="_p  ->* [] { return relational_operator::equal; }
    | "!="_p  ->* [] { return relational_operator::not_equal; }
    | "<="_p  ->* [] { return relational_operator::less_equal; }
    | "<"_p   ->* [] { return relational_operator::less; }
    | ">="_p  ->* [] { return relational_operator::greater_equal; }
    | ">"_p   ->* [] { return relational_operator::greater; }
    | "in"_p  ->* [] { return relational_operator::in; }
    | "!in"_p ->* [] { return relational_operator::not_in; }
    | "ni"_p  ->* [] { return relational_operator::ni; }
    | "!ni"_p ->* [] { return relational_operator::not_ni; }
    | "[+"_p  ->* [] { return relational_operator::in; }
    | "[-"_p  ->* [] { return relational_operator::not_in; }
    | "+]"_p  ->* [] { return relational_operator::ni; }
    | "-]"_p  ->* [] { return relational_operator::not_ni; }
    ;
  auto ws = ignore(*parsers::space);
  auto pred
    = (operand >> ws >> operation >> ws >> operand) ->* to_predicate
    ;
  return pred;
  // clang-format on
}

} // namespace

template <class Iterator>
bool predicate_parser::parse(Iterator& f, const Iterator& l,
                             unused_type) const {
  static auto p = make_predicate_parser();
  return p(f, l, unused);
}

template <class Iterator>
bool predicate_parser::parse(Iterator& f, const Iterator& l,
                             predicate& a) const {
  using namespace parsers;
  static auto p = make_predicate_parser();
  return p(f, l, a);
}

template bool
predicate_parser::parse(std::string::iterator&, const std::string::iterator&,
                        unused_type) const;
template bool
predicate_parser::parse(std::string::iterator&, const std::string::iterator&,
                        predicate&) const;

template bool
predicate_parser::parse(std::string::const_iterator&,
                        const std::string::const_iterator&, unused_type) const;
template bool
predicate_parser::parse(std::string::const_iterator&,
                        const std::string::const_iterator&, predicate&) const;

template bool
predicate_parser::parse(char const*&, char const* const&, unused_type) const;
template bool
predicate_parser::parse(char const*&, char const* const&, predicate&) const;

namespace {

template <class Iterator>
static auto make_expression_parser() {
  using namespace parser_literals;
  using chain = std::vector<std::tuple<bool_operator, expression>>;
  using raw_expr = std::tuple<expression, chain>;
  // Converts a "raw" chain of sub-expressions and transforms it into an
  // expression tree.
  auto to_expr = [](raw_expr expr) -> expression {
    auto& [x, xs] = expr;
    if (xs.empty())
      return x;
    // We split the expression chain at each OR node in order to take care of
    // operator precedance: AND binds stronger than OR.
    disjunction dis;
    auto con = conjunction{x};
    for (auto& [op, expr] : xs)
      if (op == bool_operator::logical_and) {
        con.emplace_back(std::move(expr));
      } else if (op == bool_operator::logical_or) {
        VAST_ASSERT(!con.empty());
        if (con.size() == 1)
          dis.emplace_back(std::move(con[0]));
        else
          dis.emplace_back(std::move(con));
        con = conjunction{std::move(expr)};
      } else {
        VAST_ASSERT(!"negations must not exist here");
      }
    if (con.size() == 1)
      dis.emplace_back(std::move(con[0]));
    else
      dis.emplace_back(std::move(con));
    return dis.size() == 1 ? std::move(dis[0]) : expression{dis};
  };
  auto ws = ignore(*parsers::space);
  auto negate_expr = [](expression expr) { return negation{std::move(expr)}; };
  rule<Iterator, expression> expr;
  rule<Iterator, expression> group;
  // clang-format off
  auto expand = [](data x) {
    auto pred = to_value_predicate(std::move(x));
    return caf::visit(expander{}, expression{std::move(pred)});
  };
  auto pred_expr
    = parsers::predicate ->* [](predicate p) { return expression{std::move(p)}; }
    | parsers::data ->* expand;
    ;
  group
    = '(' >> ws >> ref(expr) >> ws >> ')'
    | '!' >> ws >> pred_expr ->* negate_expr
    | '!' >> ws >> '(' >> ws >> (ref(expr) ->* negate_expr) >> ws >> ')'
    | pred_expr
    ;
  auto and_or
    = "||"_p  ->* [] { return bool_operator::logical_or; }
    | "&&"_p  ->* [] { return bool_operator::logical_and; }
    ;
  expr
    // One embedding of the group rule is intentionally not wrapped in a
    // rule_ref, because otherwise the reference count of the internal
    // shared_ptr would go to 0 at end of scope. We don't need this
    // precaution for the expr rule because that is part of the return
    // expression.
    = (group >> *(ws >> and_or >> ws >> ref(group)) >> ws) ->* to_expr
    ;
  // clang-format on
  return expr >> parsers::eoi;
}

} // namespace

template <class Iterator, class Attribute>
bool expression_parser::parse(Iterator& f, const Iterator& l,
                              Attribute& x) const {
  static auto p = make_expression_parser<Iterator>();
  return p(f, l, x);
}

template bool
expression_parser::parse(std::string::iterator&, const std::string::iterator&,
                         unused_type&) const;
template bool
expression_parser::parse(std::string::iterator&, const std::string::iterator&,
                         expression&) const;

template bool expression_parser::parse(std::string::const_iterator&,
                                       const std::string::const_iterator&,
                                       unused_type&) const;
template bool
expression_parser::parse(std::string::const_iterator&,
                         const std::string::const_iterator&, expression&) const;

template bool
expression_parser::parse(char const*&, char const* const&, unused_type&) const;
template bool
expression_parser::parse(char const*&, char const* const&, expression&) const;

} // namespace vast
